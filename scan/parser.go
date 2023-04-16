package scan

import (
	"github.com/getkin/kin-openapi/openapi3"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Parser is a struct that holds the state of the parser.
type Parser struct {
	fileSet        *token.FileSet
	spec           *openapi3.T
	packageName    string
	file           *ast.File
	logger         *Logger
	structComments map[string]*structComment
	fieldComment   map[string]*fieldComment
	typeMap        map[string]*ast.TypeSpec
	//structs        map[string]*ast.TypeSpec
	schemaMap  map[string]*openapi3.Schema
	operations []*openAPIOperation
	queue      map[string]*ast.TypeSpec

	fieldMap  map[string]*ast.Field
	structMap map[string]*ast.StructType
	meta      string

	//interfaces        map[string]*ast.TypeSpec
}

// NewParser creates a new instance of the Parser struct.
func NewParser(logger *Logger) *Parser {
	return &Parser{
		logger:  logger,
		fileSet: token.NewFileSet(),
		spec: &openapi3.T{
			OpenAPI: "3.0.0",
			Info: &openapi3.Info{
				Title:       "Sample Spec - OpenAPI 3.1",
				Version:     "3.1.0",
				Description: "Sample description.",
			},
			Servers: openapi3.Servers{
				&openapi3.Server{URL: "http://localhost:8080"},
			},
			Paths: map[string]*openapi3.PathItem{},
			Components: &openapi3.Components{
				Schemas: map[string]*openapi3.SchemaRef{},
			},
		},
		fieldComment:   map[string]*fieldComment{},
		structComments: map[string]*structComment{},
		typeMap:        make(map[string]*ast.TypeSpec),
		schemaMap:      map[string]*openapi3.Schema{},
		operations:     []*openAPIOperation{},
		queue:          map[string]*ast.TypeSpec{},
		fieldMap:       map[string]*ast.Field{},
		structMap:      map[string]*ast.StructType{},
		//structs:        map[string]*ast.TypeSpec{},
	}
}

func (p *Parser) WithMetaPath(path string) *Parser {
	p.meta = strings.TrimPrefix(path, "/")
	return p
}

func (p *Parser) GetSpec(dir string) (*openapi3.T, error) {
	err := p.parseDir(dir)
	for _, ts := range p.typeMap {
		p.createOpenAPISchema("", ts)
	}
	for _, op := range p.operations {
		p.generateOperation(op)
	}

	if len(p.meta) > 1 {

	}
	return p.spec, err
}

func (p *Parser) parseDir(dir string) error {
	return filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip directories
			p.logger.Info("Processing directory: %s\n", filePath)
			return nil
		}

		// Parse only .go files
		if filepath.Ext(filePath) != ".go" {
			return nil
		}

		// Parse the file
		file, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		p.file = file

		// Iterate through the comments in the file
		for _, comment := range file.Comments {
			if comment.Pos() < file.Package {
				switch len(p.meta) {
				case 0:
					p.extractOpenAPIInfo(comment)
					break
				default:
					if p.meta == filePath[len(dir)+1:] {
						p.extractOpenAPIInfo(comment)
						break
					}
				}
			}
		}

		// Process the file
		return p.ProcessFile(filePath[len(dir)+1:], file)
	})
}

func (p *Parser) ProcessFile(path string, file *ast.File) error {
	p.logger.Debug("processing definitions in file: %s", path)
	// Store the package name
	p.packageName = file.Name.Name

	// Traverse the AST to find structs, methods, and interfaces
	for _, decl := range file.Decls {
		switch declType := decl.(type) {
		case *ast.GenDecl:
			switch declType.Tok {
			case token.TYPE:
				// Handle type declarations
				for _, spec := range declType.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if _, ok := p.typeMap[ts.Name.Name]; ok {
							p.logger.Warn("duplicate struct `%s` are not supported", ts.Name.Name)
							continue
						}

						switch ts.Type.(type) {
						case *ast.StructType:

							t, ok := ts.Type.(*ast.StructType)
							if !ok {
								break
							}
							p.extractStructComments(ts.Name.Name, declType.Doc)

							for _, field := range t.Fields.List {
								if field == nil || field.Names == nil && len(field.Names) == 0 {
									p.logger.Debug("fields not found for %s", ts.Name.Name)
									continue
								}
								key := filepath.Join(ts.Name.Name, field.Names[0].Name)
								if field.Doc == nil {
									p.logger.Debug("openapi annotations not found for %s", key)
									continue
								}
								p.extractFieldComments(key, field.Doc)
								p.fieldMap[field.Names[0].Name] = field
							}

							p.typeMap[ts.Name.Name] = ts
							p.structMap[ts.Name.Name] = t
						case *ast.InterfaceType:
							iface, ok := ts.Type.(*ast.InterfaceType)
							if !ok {
								break
							}

							for _, field := range iface.Methods.List {
								if field == nil || field.Names == nil && len(field.Names) == 0 {
									p.logger.Debug("fields not found for %s", ts.Name.Name)
									continue
								}
								key := filepath.Join(ts.Name.Name, field.Names[0].Name)
								if field.Doc == nil {
									p.logger.Debug("openapi annotations not found for %s", key)
									continue
								}
								openAPIOp, err := extractOpenAPIOperation(key, field.Doc)
								if err != nil {
									p.logger.Fatal(err.Error())
								}
								p.operations = append(p.operations, openAPIOp)
							}
						}
					}
				}
			}
		case *ast.FuncDecl:
			// Handle function declarations
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				p.logger.Info("error processing func: %s", fn.Name.Name)
				continue
			}

			openAPIOp, err := extractOpenAPIOperation(fn.Name.Name, fn.Doc)
			if err != nil {
				continue
			}
			p.operations = append(p.operations, openAPIOp)
		default:
			p.logger.Debug("not supported")
		}
	}

	return nil
}
