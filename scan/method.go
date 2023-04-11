package scan

import (
	"github.com/getkin/kin-openapi/openapi3"
	"path/filepath"
	"strconv"
	"strings"
)

type interfaceComments struct {
	Server      string
	Summary     string
	Description string
	Tags        []string
	ID          string
	Path        string
}

type methodComments struct {
	Server      string
	Summary     string
	Description string
	Tags        []string
	ID          string
	Parameters  []*parameterComments
	Success     *responseComments
	Failure     *responseComments
	Path        string
	Method      string
	Body        string
}

type parameterComments struct {
	Name        string
	In          string
	Type        string
	Required    bool
	Description string
}

type responseComments struct {
	Code        int
	Description string
	Schema      *openapi3.SchemaRef
}

// ParseRequestType extracts information about the request type from the interface method comments.
func (p *Parser) ParseRequestType(mc *methodComments) *openapi3.RequestBodyRef {
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
	//requestBody := &openapi3.RequestBody{
	//	Required: true,
	//	Content:  make(map[string]*openapi3.MediaType),
	//}
	//
	//for _, param := range mc.Parameters {
	//	if param.In == "query" {
	//		requestBody.Content["application/x-www-form-urlencoded"] = &openapi3.MediaType{
	//			Schema: &openapi3.SchemaRef{
	//				Value: &openapi3.Schema{
	//					Type: param.Type,
	//				},
	//			},
	//		}
	//	} else if param.In == "header" {
	//		requestBody.Content["application/json"] = &openapi3.MediaType{
	//			Schema: &openapi3.SchemaRef{
	//				Value: &openapi3.Schema{
	//					Type: param.Type,
	//				},
	//			},
	//		}
	//	} else if param.In == "path" {
	//		if !strings.Contains(mc.Path, "{"+param.Name+"}") {
	//			return nil, fmt.Errorf("path parameter '%s' not found in path '%s'", param.Name, mc.Path)
	//		}
	//		requestBody.Content["application/json"] = &openapi3.MediaType{
	//			Schema: &openapi3.SchemaRef{
	//				Value: &openapi3.Schema{
	//					Type: param.Type,
	//				},
	//			},
	//		}
	//	}
	//}
	//
	//return &openapi3.RequestBodyRef{
	//	//Ref:   filepath.Join("#/components/schemas", mc.Req, "/"),
	//	Value: requestBody,
	//}, nil
}

// ParseResponseType extracts information about the response type from the interface method comments.
func (p *Parser) ParseResponseType(responseComments *responseComments) (*openapi3.ResponseRef, error) {
	if responseComments == nil {
		return nil, nil
	}

	response := &openapi3.Response{
		Description: &responseComments.Description,
		Content:     make(map[string]*openapi3.MediaType),
	}

	contentType := "application/json"
	//TODO: Support content
	//if methodComments.Success.Content != "" {
	//	contentType = methodComments.Success.Content
	//}

	// Support custom types
	response.Content[contentType] = &openapi3.MediaType{
		Schema: &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: "object",
			},
		},
	}

	return &openapi3.ResponseRef{
		Value: response,
		Ref:   responseComments.Schema.Ref,
	}, nil
}

func parseParamComment(comment string) *parameterComments {
	// Example format: @param name in type required "description"
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

func parseResponseComment(comment string, structs map[string]struct{}) *responseComments {
	// Example format: @success 200 {object} SampleResponse
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

	return &responseComments{
		Code:        code,
		Description: strings.Join(parts[0:], " "),
		Schema: &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: "object",
			},
			Ref: filepath.Join("#/components/schemas", schema, "/"),
		},
	}
}

func getInterfaceComments(comments []string) *interfaceComments {
	ic := &interfaceComments{}
	for _, c := range comments {
		text := strings.TrimSpace(strings.TrimLeft(c, "/"))
		if strings.HasPrefix(text, "@tags") {
			ic.Tags = parseTagsComment(text)
		} else if strings.HasPrefix(text, "@summary") {
			ic.Summary = strings.TrimSpace(strings.TrimPrefix(text, "@summary"))
		} else if strings.HasPrefix(text, "@description") {
			ic.Description = strings.TrimSpace(strings.TrimPrefix(text, "@description"))
		} else if strings.HasPrefix(text, "@id") {
			ic.ID = strings.TrimSpace(strings.TrimPrefix(text, "@id"))
		} else if strings.HasPrefix(text, "@server") {
			ic.Server = strings.TrimSpace(strings.TrimPrefix(text, "@server"))
		} else if strings.HasPrefix(text, "@path") {
			ic.Path = strings.TrimSpace(strings.TrimPrefix(text, "@path"))
		}
	}

	return ic
}

func (p *Parser) getMethodComments(comments []string) *methodComments {
	mc := &methodComments{}
	for _, c := range comments {
		text := strings.TrimSpace(strings.TrimLeft(c, "/"))
		if strings.HasPrefix(text, "@param") {
			p := parseParamComment(text)
			mc.Parameters = append(mc.Parameters, p)
		} else if strings.HasPrefix(text, "@success") {
			mc.Success = parseResponseComment(text, p.structs)
		} else if strings.HasPrefix(text, "@failure") {
			mc.Failure = parseResponseComment(text, nil)
		} else if strings.HasPrefix(text, "@tags") {
			mc.Tags = parseTagsComment(text)
		} else if strings.HasPrefix(text, "@summary") {
			mc.Summary = strings.TrimSpace(strings.TrimPrefix(text, "@summary"))
		} else if strings.HasPrefix(text, "@description") {
			mc.Description = strings.TrimSpace(strings.TrimPrefix(text, "@description"))
		} else if strings.HasPrefix(text, "@id") {
			mc.ID = strings.TrimSpace(strings.TrimPrefix(text, "@id"))
		} else if strings.HasPrefix(text, "@server") {
			mc.Server = strings.TrimSpace(strings.TrimPrefix(text, "@server"))
		} else if strings.HasPrefix(text, "@path") {
			mc.Path = strings.TrimSpace(strings.TrimPrefix(text, "@path"))
		} else if strings.HasPrefix(text, "@method") {
			mc.Method = strings.TrimSpace(strings.TrimPrefix(text, "@method"))
		} else if strings.HasPrefix(text, "@body") {
			mc.Body = strings.TrimSpace(strings.TrimPrefix(text, "@body"))
		}
	}

	return mc
}

func parseTagsComment(comment string) []string {
	// Example format: @tags tag1, tag2
	parts := strings.Fields(comment)
	if len(parts) < 2 {
		return nil
	}

	return strings.Split(parts[1], ",")
}
