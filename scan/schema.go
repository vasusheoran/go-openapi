package scan

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"go/ast"
	"reflect"
	"strings"
)

type structComment struct {
	Schema bool
	XML    xml
}

type fieldComment struct {
	Description string
	Example     string
	Deprecated  bool
	Nullable    bool
	Format      string
	Default     string
	Enum        []interface{}
	OneOf       []string
}

type xml struct {
	Name string
}

func (p *Parser) createOpenAPISchema(key string, ts *ast.TypeSpec) *openapi3.Schema {
	if sc, ok := p.schemaMap[ts.Name.Name]; ok {
		return sc
	}

	p.logger.Info("creating schema for %s", ts.Name.Name)

	structType := ts.Type.(*ast.StructType)
	if structType.Fields == nil || len(structType.Fields.List) == 0 {
		// If the struct has no fields, there's nothing to do.
		return nil
	}

	sc, ok := p.structComments[getKey("", ts.Name.Name, "")]
	if !ok || sc == nil || !sc.Schema {
		p.logger.Warn("openapi:schema not found for %s", ts.Name.Name)
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
		k := getKey("", ts.Name.Name, field.Names[0].Name)
		fc, ok := p.fieldComment[k]
		if !ok {
			p.logger.Warn("no openapi tags found for %s/%s", ts.Name.Name, field.Names[0].Name)
		}

		fieldSchemaRef, jsonTag := p.createFieldSchema(ts.Name.Name, fc, field, required)
		if len(jsonTag) == 0 {
			p.logger.Warn("no json tags found for %s/%s", ts.Name.Name, field.Names[0].Name)
			continue
		}

		schema.Properties[jsonTag] = fieldSchemaRef

		schema.Type = getOpenAPIFieldType(ts.Type)

		if structField, ok := field.Type.(*ast.StructType); ok {
			nestedSchema := p.createOpenAPISchema(ts.Name.Name, &ast.TypeSpec{Name: &ast.Ident{Name: ""}, Type: structField})
			if nestedSchema != nil {
				schema.Properties[jsonTag].Value = nestedSchema
			}
		}

		// Update the type map with the schema for the field's type.
		p.typeMap[ts.Name.Name] = ts
	}

	if len(sc.XML.Name) != 0 {
		schema.XML = &openapi3.XML{Name: sc.XML.Name}
	}

	schema.Required = required
	p.schemaMap[ts.Name.Name] = schema
	p.spec.Components.Schemas[ts.Name.Name] = &openapi3.SchemaRef{Value: schema}
	return schema
}

func (p *Parser) createFieldSchema(name string, fc *fieldComment, field *ast.Field, required []string) (*openapi3.SchemaRef, string) {

	// Parse the type of the field into an OpenAPI schema.
	//p.logger.Info("Parsing type expression for %s/%s ", cwd, field.Names[0].Name)
	fieldSchemaRef := p.ParseTypeExpr(name, field.Type)
	if fieldSchemaRef == nil {
		// If the field type cannot be parsed, skip it.
		return nil, ""
	}

	// Parse the JSON tag to get the field name and options.
	tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1]).Get("json")
	jsonTag, opts := parseJSONTag(tag)
	if jsonTag == "" {
		// Skip fields without JSON tags
		p.logger.Info("json tag not found for field %s", field.Names[0].Name)
		return nil, ""
	}

	// If the field is required, add it to the list of required fields.
	if opts.Contains("required") {
		required = append(required, jsonTag)
	}

	// Parse field and assign type to field SchemaRef
	fieldSchemaRef.Value.Type = getOpenAPIFieldType(field.Type)

	if len(fc.Example) > 0 {
		fieldSchemaRef.Value.Example = fc.Example
	}
	if len(fc.Description) > 0 {
		fieldSchemaRef.Value.Description = fc.Description
	}
	if len(fc.Format) > 0 {
		fieldSchemaRef.Value.Format = fc.Format
	}
	if len(fc.Default) > 0 {
		fieldSchemaRef.Value.Default = fc.Default
	}
	if len(fc.Enum) > 0 {
		fieldSchemaRef.Value.Enum = fc.Enum
	}
	fieldSchemaRef.Value.Nullable = fc.Nullable
	fieldSchemaRef.Value.Deprecated = fc.Deprecated
	fieldSchemaRef.Value.Enum = fc.Enum

	return fieldSchemaRef, jsonTag
}

func (p *Parser) GetTypeSpec(t ast.Expr) *ast.TypeSpec {
	switch t := t.(type) {
	case *ast.StarExpr:
		return p.GetTypeSpec(t.X)
	case *ast.Ident:
		// Handle nested cases
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

// ParseTypeExpr returns the OpenAPI schema for the given Go type expression.
// Returns nil if the expression is not a valid type.
func (p *Parser) ParseTypeExpr(key string, expr ast.Expr) *openapi3.SchemaRef {
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
			return &openapi3.SchemaRef{
				Ref:   fmt.Sprintf("#/components/schemas/%s", ts.Name.Name),
				Value: p.createOpenAPISchema(key, ts),
			}
		}
	case *ast.ArrayType:
		itemsSchemaRef := p.ParseTypeExpr(key, t.Elt)
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

func (p *Parser) extractStructComments(name string, cg *ast.CommentGroup) {
	if cg == nil {
		p.logger.Debug("no comments found for %s", name)
		return
	}
	c := &structComment{}
	for _, comment := range cg.List {
		text := strings.TrimSpace(strings.TrimLeft(comment.Text, "/"))
		if strings.Contains(text, "openapi:schema") {
			c.Schema = true
		} else if strings.HasPrefix(text, "openapi:xml") {
			c.XML.Name = strings.Trim(strings.TrimSpace(strings.TrimPrefix(text, "openapi:xml")), "\"")
		}
	}
	p.structComments[name] = c
}

func (p *Parser) extractFieldComments(name string, cg *ast.CommentGroup) {
	if cg == nil {
		p.logger.Debug("no comments found for %s", name)
		return
	}
	c := &fieldComment{}
	for _, comment := range cg.List {
		text := strings.TrimSpace(strings.TrimLeft(comment.Text, "/"))
		if strings.HasPrefix(text, "openapi:description") {
			c.Description = strings.Trim(strings.TrimSpace(strings.TrimPrefix(text, "openapi:description")), "\"")
		} else if strings.HasPrefix(text, "openapi:example") {
			c.Example = strings.Trim(strings.TrimSpace(strings.TrimPrefix(text, "openapi:example")), "\"")
		} else if strings.Contains(text, "openapi:deprecated") {
			c.Deprecated = true
		} else if strings.HasPrefix(text, "openapi:operationID") {
			c.Nullable = true
		} else if strings.HasPrefix(text, "openapi:format") {
			c.Format = strings.Trim(strings.TrimSpace(strings.TrimPrefix(text, "openapi:format")), "\"")
		} else if strings.HasPrefix(text, "openapi:default") {
			c.Default = strings.Trim(strings.TrimSpace(strings.TrimPrefix(text, "openapi:default")), "\"")
		} else if strings.HasPrefix(text, "openapi:enum") {
			enums := strings.Split(strings.Trim(strings.TrimSpace(strings.TrimPrefix(text, "openapi:enum")), "\""), ",")
			for _, enum := range enums {
				c.Enum = append(c.Enum, enum)
			}
		} else if strings.HasPrefix(text, "openapi:oneOf") {
			oneOfs := strings.Split(strings.Trim(strings.TrimSpace(strings.TrimPrefix(text, "openapi:oneOf")), "\""), " ")
			for _, oneOfType := range oneOfs {
				c.OneOf = append(c.OneOf, oneOfType)
			}
		}
	}
	p.fieldComment[name] = c
}
