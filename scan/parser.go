package scan

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
)

// Parser is a struct that holds the state of the parser.
type Parser struct {
	fileSet        *token.FileSet
	spec           *openapi3.T
	CurrentPath    string
	CurrentOp      string
	CurrentContent *openapi3.MediaType
	typeMap        map[string]*ast.TypeSpec
	file           *ast.File
	paths          map[string]*openapi3.PathItem
	comments       map[string][]string
	structs        map[string]struct{}
}

// NewParser creates a new instance of the Parser struct.
func NewParser() *Parser {
	return &Parser{
		fileSet: token.NewFileSet(),
		spec: &openapi3.T{
			OpenAPI: "3.0.0",
			Info: &openapi3.Info{
				Title:   "My API",
				Version: "1.0.0",
			},
			Servers: openapi3.Servers{
				&openapi3.Server{URL: "http://localhost:8080"},
			},
			Paths: map[string]*openapi3.PathItem{},
			Components: &openapi3.Components{
				Schemas: map[string]*openapi3.SchemaRef{},
			},
		},
		paths:    map[string]*openapi3.PathItem{},
		typeMap:  make(map[string]*ast.TypeSpec),
		comments: map[string][]string{},
		structs:  map[string]struct{}{},
	}
}

func (p *Parser) GetSpec() *openapi3.T {
	return p.spec
}

// ParseFile parses a Go source code file and updates the OpenAPI specs based on the comments in the file.
func (p *Parser) ParseFile(filename string) error {
	file, err := parser.ParseFile(p.fileSet, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	p.file = file

	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			switch decl.Tok {
			case token.TYPE:
				// Handle type declarations
				for _, spec := range decl.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						//intfDecl := decl.(*ast.GenDecl)
						for _, c := range decl.Doc.List {
							list, ok := p.comments[ts.Name.Name]
							if !ok {
								list = []string{}
							}
							list = append(list, c.Text)
							p.comments[ts.Name.Name] = list
						}

						switch ts.Type.(type) {
						case *ast.StructType:
							p.structs[ts.Name.Name] = struct{}{}
							p.ParseStructType(ts)
						case *ast.InterfaceType:
							iface, ok := ts.Type.(*ast.InterfaceType)
							if !ok {
								break
							}

							for _, field := range iface.Methods.List {
								for _, c := range field.Doc.List {
									list, ok := p.comments[field.Names[0].Name]
									if !ok {
										list = []string{}
									}
									list = append(list, c.Text)
									p.comments[field.Names[0].Name] = list
								}
							}

							p.ParseInterfaceType(ts)
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
		}
	}

	return nil
}

// ParseTypeExpr returns the OpenAPI schema for the given Go type expression.
// Returns nil if the expression is not a valid type.
func (p *Parser) ParseTypeExpr(expr ast.Expr) *openapi3.SchemaRef {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}}
		case "int", "int32", "int64":
			return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "integer", Format: t.Name}}
		case "float32", "float64":
			return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "number", Format: t.Name}}
		case "bool":
			return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "boolean"}}
		case "[]byte":
			return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "byte"}}
		case "time.Time":
			return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "date-time"}}
		default:
			ts := p.GetTypeSpec(t)
			if ts != nil {
				return &openapi3.SchemaRef{
					Ref:   fmt.Sprintf("#/components/schemas/%s", ts.Name.Name),
					Value: p.ParseStructType(ts),
				}
			} else {
				return &openapi3.SchemaRef{
					Ref:   fmt.Sprintf("#/components/schemas/%s", t.Name),
					Value: p.ParseStructType(ts),
				}
			}
			return nil
		}
	case *ast.ArrayType:
		itemsSchemaRef := p.ParseTypeExpr(t.Elt)
		if itemsSchemaRef != nil {
			return &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:  "array",
					Items: itemsSchemaRef,
				},
			}
		}
		return nil
	}

	return nil
}

func (p *Parser) GetTypeSpec(t ast.Expr) *ast.TypeSpec {
	switch t := t.(type) {
	case *ast.StarExpr:
		return p.GetTypeSpec(t.X)
	case *ast.Ident:
		// Look up cache, return value if present
		if ts, ok := p.typeMap[t.Name]; ok {
			return ts
		}
		// Traverse the package AST to find the type declaration
		var ts *ast.TypeSpec
		ast.Inspect(p.file, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.TypeSpec:
				if n.Name.Name == t.Name {
					ts = n
					return false // Stop traversal
				}
			}
			return true // Continue traversal
		})

		// Update cache
		p.typeMap[t.Name] = ts
		return ts
	}
	return nil
}

//	func (p *Parser) GetTypeSpec(t ast.Expr) *ast.TypeSpec {
//		switch t := t.(type) {
//		case *ast.StarExpr:
//			return p.GetTypeSpec(t.X)
//		case *ast.Ident:
//			if ts, ok := p.typeMap[t.Name]; ok {
//				return ts
//			}
//		}
//		return nil
//	}
func (p *Parser) ParseStructType(ts *ast.TypeSpec) *openapi3.Schema {
	structType := ts.Type.(*ast.StructType)
	if structType.Fields == nil || len(structType.Fields.List) == 0 {
		// If the struct has no fields, there's nothing to do.
		return nil
	}

	schema := &openapi3.Schema{
		Type:       "object",
		Properties: map[string]*openapi3.SchemaRef{},
	}
	required := []string{}

	for _, field := range structType.Fields.List {
		if field.Tag == nil {
			// If the field has no tag, skip it.
			continue
		}

		// Get the name and type of the field.
		//fieldName := field.Names[0].Name
		fieldType := field.Type

		// Parse the type of the field into an OpenAPI schema.
		fieldSchemaRef := p.ParseTypeExpr(fieldType)
		if fieldSchemaRef == nil {
			// If the field type cannot be parsed, skip it.
			continue
		}

		// Parse the JSON tag to get the field name and options.
		tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1]).Get("json")
		jsonTag, opts := parseJSONTag(tag)
		if jsonTag == "" {
			// Skip fields without JSON tags.
			continue
		}

		// Add the field's schema to the struct schema using the JSON tag as the property name.
		schema.Properties[jsonTag] = fieldSchemaRef

		// If the field is required, add it to the list of required fields.
		if opts.Contains("required") {
			required = append(required, jsonTag)
		}

		// Update the type map with the schema for the field's type.
		p.typeMap[ts.Name.Name] = ts
	}

	schema.Required = required
	// Update the OpenAPI specs with the struct schema.
	p.spec.Components.Schemas[ts.Name.Name] = &openapi3.SchemaRef{Value: schema}
	return schema
}
