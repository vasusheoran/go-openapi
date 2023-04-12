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
				Title:       "My API",
				Version:     "1.0.0",
				Description: "This is a sample Pet Store Server based on the OpenAPI 3.1 specification.  You can find out more about\nSwagger at [https://swagger.io](https://swagger.io). In the third iteration of the pet store, we've switched to the design first approach!\nYou can now help us improve the API whether it's by making changes to the definition itself or to the code.\nThat way, with time, we can improve the API in general, and expose some of the new features in OAS3.\n\nSome useful links:\n- [The Pet Store repository](https://github.com/swagger-api/swagger-petstore)\n- [The source API definition for the Pet Store](https://github.com/swagger-api/swagger-petstore/blob/master/src/main/resources/openapi.yaml)",
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
			case token.IMPORT:
				// Handle import declarations
			case token.CONST, token.VAR:
				// Handle const and var declarations
			}
		case *ast.FuncDecl:
			// Handle function declarations
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				p.logger.Info("error processing func: %s", fn.Name.Name)
				continue
			}
			if _, ok := p.methods[fn.Name.Name]; ok {
				return fmt.Errorf("duplicate methods `%s` are not supported", fn.Name.Name)
			}
			p.parseCommentGroup(fn.Name.Name, fn.Doc)
			p.methods[fn.Name.Name] = fn
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

//
//// parseFile parses a Go source code file and updates the OpenAPI specs based on the comments in the file.
//func (p *Parser) parseFile(filename string) error {
//	file, err := parser.ParseFile(p.fileSet, filename, nil, parser.ParseComments)
//	if err != nil {
//		return err
//	}
//
//	p.file = file
//
//	for _, decl := range file.Decls {
//		switch decl := decl.(type) {
//		case *ast.GenDecl:
//			switch decl.Tok {
//			case token.TYPE:
//				// Handle type declarations
//				for _, spec := range decl.Specs {
//					if ts, ok := spec.(*ast.TypeSpec); ok {
//						//intfDecl := decl.(*ast.GenDecl)
//						for _, c := range decl.Doc.List {
//							list, ok := p.comments[ts.Name.Name]
//							if !ok {
//								list = []string{}
//							}
//							list = append(list, c.Text)
//							p.comments[ts.Name.Name] = list
//						}
//
//						switch ts.Type.(type) {
//						case *ast.StructType:
//							//p.structs[ts.Name.Name] = struct{}{}
//							p.ParseStructType("", ts)
//						case *ast.InterfaceType:
//							iface, ok := ts.Type.(*ast.InterfaceType)
//							if !ok {
//								break
//							}
//
//							for _, field := range iface.Methods.List {
//								for _, c := range field.Doc.List {
//									list, ok := p.comments[field.Names[0].Name]
//									if !ok {
//										list = []string{}
//									}
//									list = append(list, c.Text)
//									p.comments[field.Names[0].Name] = list
//								}
//							}
//
//							p.ParseInterfaceType(ts)
//						}
//					}
//				}
//			case token.IMPORT:
//				// Handle import declarations
//			case token.CONST, token.VAR:
//				// Handle const and var declarations
//			}
//		case *ast.FuncDecl:
//			// Handle function declarations
//		}
//	}
//
//	return nil
//}
