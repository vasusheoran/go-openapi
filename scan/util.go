package scan

import (
	"github.com/getkin/kin-openapi/openapi3"
	"go/ast"
	"path/filepath"
	"strings"
)

// getParametersFromMethodComments extracts information about the request params from the interface method comments.
func getParametersFromMethodComments(pc []*Parameter) openapi3.Parameters {
	var parametersRefs openapi3.Parameters
	for _, param := range pc {
		parameter := &openapi3.Parameter{
			Name:        param.Name,
			Description: param.Description,
			Required:    param.Required == "true",
			In:          param.In,
			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: param.Type,
				},
			},
		}
		parametersRefs = append(parametersRefs, &openapi3.ParameterRef{
			Value: parameter,
		})
	}
	return parametersRefs
}

// getRequestBodyFromOperation extracts information about the request type from the method comments.
func getRequestBodyFromOperation(schema *openapi3.Schema, op *openAPIOperation) *openapi3.RequestBodyRef {
	return &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().WithSchema(schema, op.Consumes).WithDescription(op.RequestBody.Description),
	}
}

// getResponseFromOperation extracts information about the response type from the method comments.
func getResponseFromOperation(schema *openapi3.Schema, op *openAPIOperation, response *ResponseBody) *openapi3.ResponseRef {
	return &openapi3.ResponseRef{
		Value: openapi3.NewResponse().WithContent(openapi3.NewContentWithSchema(schema, op.Produces)).WithDescription(response.Description),
	}
}

func extractContentFromSting(s, prefix, suffix string) (string, string) {
	// find the "[" character in the string
	start := strings.Index(s, prefix)
	if start == -1 {
		return "", s
	}

	// find the "]" character in the string
	end := strings.Index(s, suffix)
	if end == -1 {
		return "", s
	}

	// extract the content between the brackets
	content := s[start+1 : end]

	// remove the brackets and content from the string
	result := strings.Replace(s, "["+content+"]", "", 1)

	return content, result
}

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

func parseJSONTag(tag string) (string, tagOptions) {
	options := make(tagOptions)

	parts := strings.Split(tag, ",")
	name := parts[0]

	for _, part := range parts[1:] {
		if part == "omitempty" {
			options[part] = true
		}
	}

	return name, options
}

func getOpenAPIFieldType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "bool":
			return "boolean"
		case "string":
			return "string"
		case "int", "int8", "int16", "int32", "uint", "uint8", "uint16", "uint32", "float32", "float64":
			return "number"
		default:
			return "object"
		}
	case *ast.ArrayType:
		return "array"
	case *ast.MapType:
		return "object"
	case *ast.StarExpr:
		return getOpenAPIFieldType(t.X)
	default:
		return ""
	}
}

// TODO: Use to store field types
//func fieldToTypeSpec(field *ast.Field) (*ast.TypeSpec, error) {
//	ident, ok := field.Type.(*ast.Ident)
//	if !ok {
//		selectorExpr, ok := field.Type.(*ast.SelectorExpr)
//		if !ok {
//			return nil, fmt.Errorf("unsupported field type: %T", field.Type)
//		}
//		ident = selectorExpr.Sel
//	}
//
//	spec, ok := ident.Obj.Decl.(*ast.TypeSpec)
//	if !ok {
//		return nil, fmt.Errorf("unable to convert field to TypeSpec")
//	}
//
//	return spec, nil
//}
//
//func typeToFieldSpec(name string, typ types.Type) *ast.Field {
//	ident := ast.NewIdent(name)
//	fieldType := ast.ParseExpr(types.TypeString(typ, nil))
//
//	return &ast.Field{
//		Names: []*ast.Ident{ident},
//		Type:  fieldType,
//	}
//}
