package scan

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"go/ast"
	"path/filepath"
	"reflect"
)

func getKey(base, name, field string) string {
	if len(base) > 0 && len(field) > 0 {
		return filepath.Join(base, name, field)
	} else if len(base) > 0 {
		return filepath.Join(base, name)
	} else {
		return filepath.Join(name, field)
	}
	return name
}

func (p *Parser) ParseStructType(key string, ts *ast.TypeSpec) *openapi3.Schema {
	//p.logger.Info("Parsing struct %s in dir %s ", ts.Name.Name, key)
	structType := ts.Type.(*ast.StructType)
	if structType.Fields == nil || len(structType.Fields.List) == 0 {
		// If the struct has no fields, there's nothing to do.
		return nil
	}

	sc := structComment{}
	err := p.extractOpenAPIComment(ts.Name.Name, "struct", p.comments[getKey(key, ts.Name.Name, "")], &sc)
	if err != nil {
		// print error but continue to parse specs
		p.logger.Debug(err.Error())
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

		// Parse the type of the field into an OpenAPI schema.
		//p.logger.Info("Parsing type expression for %s/%s ", cwd, field.Names[0].Name)
		fieldSchemaRef := p.ParseTypeExpr(ts.Name.Name, field.Type)
		if fieldSchemaRef == nil {
			// If the field type cannot be parsed, skip it.
			continue
		}

		// Get the name and type of the field.
		fc := &fieldComment{}
		err = p.extractOpenAPIComment(key, "field", p.comments[getKey(ts.Name.Name, field.Names[0].Name, "")], fc)
		if err != nil {
			// print error but continue to parse specs
			p.logger.Error(err.Error())
			return nil
		}

		// Parse the JSON tag to get the field name and options.
		tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1]).Get("json")
		jsonTag, opts := parseJSONTag(tag)
		if jsonTag == "" {
			// Skip fields without JSON tags
			p.logger.Warn("no json tag found for field %s", key)
			continue
		}

		// Parse the openapi tags to get the description, example, and other metadata.
		tag = reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1]).Get("description")
		tagValue, opts := parseJSONTag(tag)
		if tagValue == "" {
			// If the "description" property is not found in the openapi tag,
			// check if there is an "description" comment for the field.

		}

		// Parse field and assign type to field Schema
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
		// TODO: support enum by interface
		//fieldSchemaRef.Value.Enum = fc.Enum

		schema.Properties[jsonTag] = fieldSchemaRef

		// If the field is required, add it to the list of required fields.
		if opts.Contains("required") {
			required = append(required, jsonTag)
		}

		// Handle nested struct types
		if structField, ok := field.Type.(*ast.StructType); ok {
			nestedSchema := p.ParseStructType(ts.Name.Name, &ast.TypeSpec{Name: &ast.Ident{Name: ""}, Type: structField})
			if nestedSchema != nil {
				schema.Properties[jsonTag].Value = nestedSchema
			}
		}

		// Update the type map with the schema for the field's type.
		p.typeMap[ts.Name.Name] = ts
	}

	schema.Required = required
	// Update the OpenAPI specs with the struct schema.
	p.spec.Components.Schemas[ts.Name.Name] = &openapi3.SchemaRef{Value: schema}
	return schema
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
				Value: p.ParseStructType(key, ts),
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
