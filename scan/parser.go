package scan

import (
	"fmt"
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
	fileSet     *token.FileSet
	spec        *openapi3.T
	packageName string
	file        *ast.File
	logger      *Logger
	comments    map[string][]string
	typeMap     map[string]*ast.TypeSpec
	structs     map[string]*ast.TypeSpec
	interfaces  map[string]*ast.TypeSpec
	methods     map[string]*ast.FuncDecl
	paths       map[string]*openapi3.PathItem
	schemaMap   map[string]*openapi3.Schema
	operations  map[string]*openapi3.Operation
}

// NewParser creates a new instance of the Parser struct.
func NewParser() *Parser {
	return &Parser{
		logger:  NewLogger(""),
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
		comments:   map[string][]string{},
		paths:      map[string]*openapi3.PathItem{},
		typeMap:    make(map[string]*ast.TypeSpec),
		schemaMap:  map[string]*openapi3.Schema{},
		operations: map[string]*openapi3.Operation{},
		structs:    map[string]*ast.TypeSpec{},
		methods:    map[string]*ast.FuncDecl{},
		interfaces: map[string]*ast.TypeSpec{},
	}
}

func (p Parser) WithLogLevel(level string) {
	p.logger.level = GetLogLevel(level)
}

func (p *Parser) GetSpec(dir string) (*openapi3.T, error) {
	err := p.parseDir(dir)
	for _, spec := range p.structs {
		p.ParseStructType("", spec)
	}
	for _, spec := range p.interfaces {
		p.ParseInterfaceType(spec)
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

		// Process the file
		return p.ProcessFile(file)
	})
}

func (p *Parser) ProcessFile(file *ast.File) error {
	p.logger.Info("processing package: %s", file.Name.Name)
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

						switch ts.Type.(type) {
						case *ast.StructType:
							if _, ok := p.structs[ts.Name.Name]; ok {
								return fmt.Errorf("duplicate struct `%s` are not supported", ts.Name.Name)
							}

							t, ok := ts.Type.(*ast.StructType)
							if !ok {
								break
							}
							p.parseCommentGroup(ts.Name.Name, declType.Doc)

							for _, field := range t.Fields.List {
								key := filepath.Join(ts.Name.Name, field.Names[0].Name)
								if field.Doc == nil {
									p.logger.Warn("no openapi tag found for %s", key)
									continue
								}
								p.parseCommentGroup(key, field.Doc)
							}

							p.structs[ts.Name.Name] = ts
						case *ast.InterfaceType:
							iface, ok := ts.Type.(*ast.InterfaceType)
							if !ok {
								break
							}

							p.parseCommentGroup(ts.Name.Name, declType.Doc)

							for _, field := range iface.Methods.List {
								key := filepath.Join(ts.Name.Name, field.Names[0].Name)
								if field.Doc == nil {
									p.logger.Warn("no openapi tag found for %s", key)
									continue
								}
								p.parseCommentGroup(key, field.Doc)
							}

							if _, ok := p.interfaces[ts.Name.Name]; ok {
								return fmt.Errorf("duplicate interfaces `%s` are not supported", ts.Name.Name)
							}
							p.interfaces[ts.Name.Name] = ts
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

			if fn.Name.Name == "main" {
				p.parseOpenAPIInfo(fn.Pos())
			}

			if _, ok := p.methods[fn.Name.Name]; ok {
				return fmt.Errorf("duplicate methods `%s` are not supported", fn.Name.Name)
			}
			p.parseCommentGroup(fn.Name.Name, fn.Doc)
			p.methods[fn.Name.Name] = fn
		default:
			p.logger.Debug("not supported")
		}
	}

	return nil
}

func (p *Parser) parseCommentGroup(name string, cg *ast.CommentGroup) {
	if cg == nil {
		p.logger.Warn("no comments found for %s", name)
		return
	}

	list, ok := p.comments[name]
	if !ok {
		list = []string{}
	}

	for _, c := range cg.List {
		if !strings.Contains(c.Text, "// openapi:") {
			continue
		}
		list = append(list, c.Text)
	}

	if len(list) == 0 {
		p.logger.Debug("no openapi tag found for %s", name)
		return
	}
	p.comments[name] = list
}

// Parse the given comments and extract OpenAPI info
func (p *Parser) parseOpenAPIInfo(pkgDeclEndPos token.Pos) {
	if p.file.Doc == nil || len(p.file.Doc.List) == 0 || p.file.Doc.Pos() > pkgDeclEndPos {
		p.logger.Debug("no openapi info comments found for file %s", p.file.Name.Name)
		return
	}

	for i := 0; i < len(p.file.Doc.List); i++ {
		text := p.file.Doc.List[i].Text
		if strings.HasPrefix(text, "// openapi:info") {
			fields := strings.Fields(text[11:])
			switch fields[0] {
			case "info":
				switch fields[1] {
				case "title": // join " " and trim \"
					p.spec.Info.Description = strings.Trim(strings.TrimPrefix(p.file.Doc.List[i].Text, "// openapi:info title "), "\"")
				case "description":
					start := i + 1
					if fields[2] == "start" {
						endIdx := -1
						for i = i + 1; i < len(p.file.Doc.List); i++ {
							if strings.Contains(p.file.Doc.List[i].Text, "// openapi:") {
								endIdx = i
								break
							}
						}
						if endIdx == -1 {
							continue
						}
						var sb strings.Builder
						for i = start; i < endIdx; i++ {
							sb.WriteString(strings.TrimSpace(p.file.Doc.List[i].Text[2:]))
							sb.WriteString("\n")
						}
						p.spec.Info.Description = sb.String()
					} else {
						p.spec.Info.Description = strings.Trim(strings.TrimPrefix(p.file.Doc.List[i].Text, "// openapi:info description "), "\"")
					}
				case "version":
					p.spec.Info.Version = strings.Trim(fields[2], "\"")
				case "oas":
					p.spec.OpenAPI = strings.Trim(fields[2], "\"")
				case "server":
					p.spec.Servers = openapi3.Servers{}
					serverList := strings.Split(strings.Trim(strings.TrimPrefix(p.file.Doc.List[i].Text, "// openapi:info server "), "\""), " ")
					for _, url := range serverList {
						s := &openapi3.Server{
							URL:         url,
							Description: "",
							Variables:   nil,
						}
						p.spec.Servers = append(p.spec.Servers, s)
					}
				}
			}
		}
	}
}
func findPackageDeclEndPos(file *ast.File) token.Pos {
	// Check if the file has a package declaration
	if file.Name == nil || file.Name.Name == "" {
		return file.Pos()
	}

	// Get the position of the package keyword
	packagePos := file.Package

	// Find the position of the next semicolon
	semicolonPos := token.Pos(0)
	for _, imp := range file.Imports {
		if semicolonPos == token.Pos(0) || imp.End() > semicolonPos {
			semicolonPos = imp.End()
		}
	}
	for _, decl := range file.Decls {
		if pos := decl.Pos(); pos > packagePos && (semicolonPos == token.Pos(0) || pos < semicolonPos) {
			semicolonPos = pos
		}
	}

	// Return the position after the semicolon
	return semicolonPos + 1
}
