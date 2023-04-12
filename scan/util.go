package scan

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"go/ast"
	"path/filepath"
	"strconv"
	"strings"
)

// extractOpenAPIComment returns the value of the first "// openapi:example" comment for the field.
// TODO: user reflect to detect typeOf object
func (p *Parser) extractOpenAPIComment(name, typeOf string, comments []string, response interface{}) error {
	switch typeOf {
	case "struct":
		res, ok := response.(*structComment)
		if !ok {
			return fmt.Errorf("failed to parse comments for %s", name)
		}
		p.extractStructComments(comments, res)
	case "field":
		res, ok := response.(*fieldComment)
		if !ok {
			return fmt.Errorf("failed to parse comments for %s", name)
		}
		p.extractFieldComments(comments, res)
	default:
		return fmt.Errorf("object type %s is unknown", typeOf)
	}
	return nil
}

func (p *Parser) extractStructComments(comments []string, sc *structComment) {
	for _, text := range comments {
		text = strings.TrimSpace(strings.TrimLeft(text, "/"))
		if strings.HasPrefix(text, "openapi:description") {
			sc.Description = strings.TrimSpace(strings.TrimPrefix(text, "openapi:description"))
		} else if strings.HasPrefix(text, "openapi:id") {
			sc.OperationID = strings.TrimSpace(strings.TrimPrefix(text, "openapi:id"))
		}
	}
}

func (p *Parser) extractFieldComments(comments []string, sc *fieldComment) {
	for _, text := range comments {
		text = strings.TrimSpace(strings.TrimLeft(text, "/"))
		if strings.HasPrefix(text, "openapi:description") {
			sc.Description = strings.Trim(strings.TrimSpace(strings.TrimPrefix(text, "openapi:description")), "\"")
		} else if strings.HasPrefix(text, "openapi:example") {
			sc.Example = strings.Trim(strings.TrimSpace(strings.TrimPrefix(text, "openapi:example")), "\"")
		} else if strings.Contains(text, "openapi:deprecated") {
			sc.Deprecated = true
		} else if strings.HasPrefix(text, "openapi:id") {
			sc.Nullable = true
		} else if strings.HasPrefix(text, "openapi:format") {
			sc.Format = strings.Trim(strings.TrimSpace(strings.TrimPrefix(text, "openapi:format")), "\"")
		} else if strings.HasPrefix(text, "openapi:default") {
			sc.Default = strings.Trim(strings.TrimSpace(strings.TrimPrefix(text, "openapi:default")), "\"")
		} else if strings.HasPrefix(text, "openapi:enum") {
			enums := strings.Split(strings.Trim(strings.TrimSpace(strings.TrimPrefix(text, "openapi:enum")), "\""), ",")
			for _, enum := range enums {
				sc.Enum = append(sc.Enum, enum)
			}
		}
	}
}

// getParametersFromMethodComments extracts information about the request params from the interface method comments.
func getParametersFromMethodComments(pc []*parameterComments) openapi3.Parameters {
	var parametersRefs openapi3.Parameters
	for _, param := range pc {
		parameter := &openapi3.Parameter{
			Name:        param.Name,
			Description: param.Description,
			Required:    param.Required,
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

// getRequestBodyFromMethodComments extracts information about the request type from the interface method comments.
func getRequestBodyFromMethodComments(mc *methodComments) *openapi3.RequestBodyRef {
	//if mc.
	if len(mc.Parameters) == 0 {
		return nil
	}

	if len(mc.Body) == 0 {
		return nil
	}

	schemaRef := openapi3.NewSchemaRef(filepath.Join("#/components/schemas/", mc.Body, "/"), nil)

	content := openapi3.NewContent()
	content["application/json"] = openapi3.NewMediaType().WithSchemaRef(schemaRef)
	return &openapi3.RequestBodyRef{Value: openapi3.NewRequestBody().WithContent(content)}
}

func parseResponseComment(comment string, structs map[string]*ast.TypeSpec) *responseComments {
	// Example format: openapi:success 200 {object} SampleResponse
	parts := strings.Fields(comment)
	if len(parts) < 3 {
		return nil
	}

	var code int
	var schema string

	code, err := strconv.Atoi(parts[1])
	if err != nil {
		panic(err)
	}

	parts = parts[2:]

	for i, part := range parts {
		if _, ok := structs[part]; ok {
			schema = part
			parts = append(parts[:i], parts[i+1:]...)
			break
		}
	}

	if len(schema) == 0 {
		return nil
	}

	schemaRef := openapi3.NewSchemaRef(filepath.Join("#/components/schemas/", schema, "/"), nil)

	return &responseComments{
		Code:        code,
		Description: strings.Join(parts[0:], " "),
		Schema:      schemaRef,
	}
}

func extractComment(comment, prefix string) string {
	if strings.HasPrefix(comment, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(comment, prefix))
	}
	return ""
}

func getInterfaceComments(comments []string) *interfaceComments {
	ic := &interfaceComments{}
	for _, c := range comments {
		text := strings.TrimSpace(strings.TrimLeft(c, "/"))
		if strings.HasPrefix(text, "openapi:description") {
			ic.Description = strings.TrimSpace(strings.TrimPrefix(text, "openapi:description"))
		} else if strings.HasPrefix(text, "openapi:name") {
			ic.Name = strings.TrimSpace(strings.TrimPrefix(text, "openapi:name"))
		} else if strings.HasPrefix(text, "openapi:server") {
			externalDocs := strings.SplitN(strings.TrimSpace(strings.TrimPrefix(text, "openapi:external-docs")), " ", 1)
			ic.ExternalDocs.URL = externalDocs[0]
			ic.ExternalDocs.Description = externalDocs[1]
		} else if strings.HasPrefix(text, "openapi:path") {
			ic.BasePath = strings.TrimSpace(strings.TrimPrefix(text, "openapi:path"))
		}
	}

	return ic
}

// getResponseFromMethodComments extracts information about the response type from the interface method comments.
func getResponseFromMethodComments(responseComments *responseComments) (*openapi3.ResponseRef, error) {
	if responseComments == nil {
		return nil, nil
	}

	response := &openapi3.Response{
		Description: &responseComments.Description,
		Content:     make(map[string]*openapi3.MediaType),
	}

	content := openapi3.NewMediaType().WithSchemaRef(responseComments.Schema)
	//TODO: Support content
	//if methodComments.Success.Content != "" {
	//	contentType = methodComments.Success.Content
	//}

	// Support custom types
	response.Content["application/json"] = content

	return &openapi3.ResponseRef{
		Value: response,
	}, nil
}

func parseParamComment(comment string) *parameterComments {
	// Example format: openapi:param name in type required "description"
	parts := strings.Fields(comment)
	if len(parts) < 5 {
		return nil
	}

	p := &parameterComments{
		Name:        parts[1],
		In:          parts[2],
		Type:        parts[3],
		Required:    parts[4] == "true",
		Description: strings.Join(parts[5:], " "),
	}

	return p
}

func (p *Parser) getMethodComments(comments []string, op string) *methodComments {
	mc := &methodComments{}
	for _, c := range comments {
		text := strings.TrimSpace(strings.TrimLeft(c, "/"))
		if strings.HasPrefix(text, "openapi:param") {
			p := parseParamComment(text)
			mc.Parameters = append(mc.Parameters, p)
		} else if strings.HasPrefix(text, "openapi:success") {
			mc.Success = parseResponseComment(text, p.structs)
		} else if strings.HasPrefix(text, "openapi:failure") {
			mc.Failure = parseResponseComment(text, p.structs)
		} else if strings.HasPrefix(text, "openapi:tags") {
			mc.Tags = parseTagsComment(text)
		} else if strings.HasPrefix(text, "openapi:summary") {
			mc.Summary = strings.TrimSpace(strings.TrimPrefix(text, "openapi:summary"))
		} else if strings.HasPrefix(text, "openapi:description") {
			mc.Description = strings.TrimSpace(strings.TrimPrefix(text, "openapi:description"))
		} else if strings.HasPrefix(text, "openapi:id") {
			mc.ID = strings.TrimSpace(strings.TrimPrefix(text, "openapi:id"))
		} else if strings.HasPrefix(text, "openapi:server") {
			mc.Server = strings.TrimSpace(strings.TrimPrefix(text, "openapi:server"))
		} else if strings.HasPrefix(text, "openapi:path") {
			mc.Path = strings.TrimSpace(strings.TrimPrefix(text, "openapi:path"))
		} else if strings.HasPrefix(text, "openapi:method") {
			mc.Method = strings.TrimSpace(strings.TrimPrefix(text, "openapi:method"))
		} else if strings.HasPrefix(text, "openapi:body") {
			mc.Body = strings.TrimSpace(strings.TrimPrefix(text, "openapi:body"))
		}
	}

	return mc
}

func parseTagsComment(comment string) []string {
	// Example format: openapi:tags tag1, tag2
	parts := strings.Fields(comment)
	if len(parts) < 2 {
		return nil
	}

	return strings.Split(parts[1], ",")
}
