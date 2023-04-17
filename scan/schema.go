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
	Name   string
	XML    xml
}

type fieldComment struct {
	Description string
	Example     string
	Deprecated  bool
	Nullable    bool
	Format      string
	Default     string
	Name        string
	Enum        []interface{}
	OneOf       []string
}

type xml struct {
	Name string
}

func (p *Parser) createOpenAPISchema(structNameInSchema string, ts *ast.TypeSpec) *openapi3.Schema {
	sc, ok := p.structComments[getKey("", structNameInSchema, "")]
	if !ok || sc == nil || !sc.Schema {
		p.logger.Debug("openapi:schema not found for %s", structNameInSchema)
		return nil
	}

	if ts == nil {
		p.logger.Fatal("schema not found for `%s`", structNameInSchema)
	}

	var schema *openapi3.Schema
	if schema, ok = p.schemaMap[structNameInSchema]; !ok {
		p.logger.Debug("creating schema for %s", structNameInSchema)

		schema = &openapi3.Schema{
			Type:       "object",
			Properties: map[string]*openapi3.SchemaRef{},
		}
	} else {
		p.logger.Debug("found schema for %s", structNameInSchema)
	}

	required := []string{}

	// Parse if nested object

	if schema.Type == "object" {

		structType := ts.Type.(*ast.StructType)
		if structType.Fields == nil || len(structType.Fields.List) == 0 {
			// If the struct has no fields, there's nothing to do.
			return nil
		}

		for _, field := range structType.Fields.List {
			if field.Tag == nil {
				// If the field has no tag, skip it.
				continue
			}
			//

			fieldName := p.extractFieldComments(structNameInSchema, field.Names[0].Name, field.Doc)
			if fieldName == nil {
				p.logger.Fatal("no openapi:name found for %s/%s", ts.Name.Name, field.Names[0].Name)
			}

			// Get the name and type of the field.
			fc, ok := p.fieldComment[*fieldName]
			if !ok {
				p.logger.Warn("no openapi tags found for %s/%s", structNameInSchema, field.Names[0].Name)
				//return nil
			}

			p.logger.Debug("parsing schema %s with field %s", structNameInSchema, field.Names[0].Name)
			fieldSchemaRef, jsonTag := p.createFieldSchema(structNameInSchema, fc, field, required)
			if len(jsonTag) == 0 {
				p.logger.Info("no json tags found for %s/%s", structNameInSchema, field.Names[0].Name)
				continue
			}

			schema.Type = getOpenAPIFieldType(ts.Type)
			schema.Properties[jsonTag] = fieldSchemaRef

			if fc != nil && len(fc.OneOf) > 0 {

				oneOfSchema := openapi3.NewOneOfSchema()
				p.logger.Debug("found openapi:oneOf")
				for _, name := range fc.OneOf {
					_, ok := p.structMap[name]
					if !ok {
						// Continue with a waring
						p.logger.Warn("oneOf field %s not found", name)
					}

					oneOfSchema.OneOf = append(oneOfSchema.OneOf, openapi3.NewSchemaRef(fmt.Sprintf("#/components/schemas/%s", name), nil))
				}
				schema.Properties[jsonTag].Value = oneOfSchema
				schema.Properties[jsonTag].Ref = ""
			} else if structField, isStructField := field.Type.(*ast.StructType); isStructField {
				nestedSchema := p.createOpenAPISchema(structNameInSchema, &ast.TypeSpec{Name: &ast.Ident{Name: ""}, Type: structField})
				if nestedSchema != nil {
					schema.Properties[jsonTag].Ref = fmt.Sprintf("#/components/schemas/%s", field.Names[0].Name)
					//schema.Properties[jsonTag].Value = nestedSchema
				}
			}
		}
	}

	if len(sc.XML.Name) != 0 {
		schema.XML = &openapi3.XML{Name: sc.XML.Name}
	}

	schema.Required = required
	p.schemaMap[structNameInSchema] = schema
	p.spec.Components.Schemas[structNameInSchema] = &openapi3.SchemaRef{Value: schema}
	return schema
}

func (p *Parser) createFieldSchema(name string, fc *fieldComment, field *ast.Field, required []string) (*openapi3.SchemaRef, string) {
	// Parse the JSON tag to get the field name and options.
	tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1]).Get("json")
	jsonTag, opts := parseJSONTag(tag)

	if jsonTag == "" {
		// Skip fields without JSON tags
		p.logger.Info("json tag not found for field %s", field.Names[0].Name)
		return nil, ""
	} else if jsonTag == "-" {
		p.logger.Info("skipped parsing for field %s with json tag `-`", field.Names[0].Name)
		return nil, ""
	}

	if sc, ok := p.schemaMap[fc.Name]; ok {
		return openapi3.NewSchemaRef("", sc), jsonTag
	}
	// Parse the type of the field into an OpenAPI schema.
	//p.logger.Info("Parsing type expression for %s/%s ", cwd, field.Names[0].Name)
	fieldSchemaRef := p.ParseTypeExpr(fc.Name, field.Type)
	if fieldSchemaRef == nil {
		// If the field type cannot be parsed, skip it.
		return openapi3.NewSchemaRef("", openapi3.NewSchema()), jsonTag
	}

	// If the field is required, add it to the list of required fields.
	if opts.Contains("required") {
		required = append(required, jsonTag)
	}

	// Parse field and assign type to field SchemaRef
	if fieldSchemaRef.Value == nil {
		fieldSchemaRef.Value = openapi3.NewSchema()
	}
	fieldSchemaRef.Value.Type = getOpenAPIFieldType(field.Type)

	if fc != nil {
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
	}

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
// TODO: should search cache based on openapi:name tag instead of field name
func (p *Parser) ParseTypeExpr(key string, expr ast.Expr) *openapi3.SchemaRef {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string", "error":
			return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}}
		case "int", "int32", "int64", "uint":
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
				Ref:   fmt.Sprintf("#/components/schemas/%s", key),
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

func (p *Parser) extractStructComments(name string, cg *ast.CommentGroup) *string {
	if cg == nil {
		p.logger.Debug("no comments found for %s", name)
		return nil
	}

	c := &structComment{}
	for _, comment := range cg.List {
		text := strings.TrimSpace(strings.TrimLeft(comment.Text, "/"))
		if strings.Contains(text, "openapi:schema") {
			c.Schema = true
			c.Name = strings.Split(strings.TrimSpace(strings.TrimPrefix(text, "openapi:schema")), " ")[0]
		} else if strings.HasPrefix(text, "openapi:xml") {
			c.XML.Name = strings.Trim(strings.TrimSpace(strings.TrimPrefix(text, "openapi:xml")), "\"")
		}
	}

	if !c.Schema {
		return nil
	}

	if len(c.Name) == 0 {
		c.Name = name
	}

	if _, ok := p.typeMap[c.Name]; ok {
		p.logger.Fatal("duplicate struct `%s` are not supported", c.Name)
	}
	p.structComments[c.Name] = c
	return &c.Name
}

func (p *Parser) extractFieldComments(schemaName string, name string, cg *ast.CommentGroup) *string {
	if cg == nil {
		cg = &ast.CommentGroup{List: []*ast.Comment{}}
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
		} else if strings.HasPrefix(text, "openapi:name") {
			c.Name = strings.Trim(strings.TrimSpace(strings.TrimPrefix(text, "openapi:name")), "\"")
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

	if len(c.Name) == 0 {
		c.Name = name
	}

	fn := fmt.Sprintf("%s/%s", schemaName, c.Name)

	p.fieldComment[fn] = c
	return &fn
}
