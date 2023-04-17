package scan

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"go/ast"
	"strings"
)

type openAPIOperation struct {
	Method      string
	OperationID string
	Path        string
	Summary     string
	Description string
	Tags        []string
	Consumes    []string
	Produces    []string
	RequestBody *RequestBody
	Responses   []*ResponseBody
	Parameters  []*Parameter
}

type RequestBody struct {
	Name        string
	Description string
}

type ResponseBody struct {
	Name        string
	Code        string
	Description string
}

type Parameter struct {
	Name        string
	In          string
	Description string
	Type        string
	Required    string
}

func (p *Parser) generateOperation(op *openAPIOperation) {
	if op == nil {
		return
	}
	p.logger.Debug("processing %s", op.OperationID)

	resp := &openapi3.Operation{}

	resp.OperationID = op.OperationID

	if op.Summary != "" {
		resp.Summary = op.Summary
	}

	if op.Description != "" {
		resp.Description = op.Description
	}

	if op.Tags != nil {
		resp.Tags = op.Tags
	}

	resp.RequestBody = getRequestBodyFromOperation(p.schemaMap[op.RequestBody.Name], op)
	resp.Parameters = getParametersFromMethodComments(op.Parameters)

	resp.Responses = make(openapi3.Responses)
	for _, responseBody := range op.Responses {
		resp.Responses[responseBody.Code] = getResponseFromOperation(p.schemaMap[responseBody.Name], op, responseBody)
	}

	// Set the op tags.
	if len(op.Tags) > 0 {
		resp.Tags = op.Tags
	}

	// Get or create the path item for the method's path.
	path := strings.Join([]string{"", op.Path}, "")
	pathItem := &openapi3.PathItem{}

	// Add the op to the path item.
	switch strings.ToUpper(op.Method) {
	case "GET":
		pathItem.Get = resp
	case "PUT":
		pathItem.Put = resp
	case "POST":
		pathItem.Post = resp
	case "DELETE":
		pathItem.Delete = resp
	case "OPTIONS":
		pathItem.Options = resp
	case "HEAD":
		pathItem.Head = resp
	case "PATCH":
		pathItem.Patch = resp
	case "TRACE":
		pathItem.Trace = resp
	case "":
		p.logger.Info("Setting default method to GET for %s", op.OperationID)
		pathItem.Get = resp
	default:
		// If the method name isn't recognized, skip it.
		p.logger.Warn("unrecognized method setting to %s", op.OperationID)
		return
	}

	p.spec.AddOperation(path, op.Method, resp)
}

func extractOpenAPIOperation(name string, cg *ast.CommentGroup) (*openAPIOperation, error) {
	op := &openAPIOperation{
		Responses:   []*ResponseBody{},
		RequestBody: &RequestBody{},
		Parameters:  []*Parameter{},
	}

	if cg == nil {
		return nil, fmt.Errorf("comments not found: %s", name)
	}

	var isValidOperation bool

	for _, comment := range cg.List {
		text := strings.TrimSpace(strings.TrimLeft(comment.Text, "//"))

		if strings.HasPrefix(text, "openapi:operation") {
			t := strings.TrimSpace(strings.TrimLeft(text, "openapi:operation"))
			parts := strings.Split(t, " ")
			if len(parts) != 3 {
				return nil, fmt.Errorf("invalid openapi:operation format: %s", name)
			}
			op.Method = parts[0]
			op.Path = parts[1]
			op.OperationID = parts[2]
			isValidOperation = true
		} else if strings.HasPrefix(text, "openapi:summary") {
			op.Summary = strings.TrimSpace(strings.TrimPrefix(text, "openapi:summary"))
		} else if strings.HasPrefix(text, "openapi:description") {
			op.Description = strings.TrimSpace(strings.TrimPrefix(text, "openapi:description"))
		} else if strings.HasPrefix(text, "openapi:tag") {
			op.Tags = append(op.Tags, strings.TrimSpace(strings.TrimPrefix(text, "openapi:tag")))
		} else if strings.HasPrefix(text, "openapi:consumes") {
			parts := strings.Split(strings.TrimSpace(strings.TrimPrefix(text, "openapi:consumes")), " ")
			op.Consumes = append(op.Consumes, parts...)
		} else if strings.HasPrefix(text, "openapi:produces") {
			parts := strings.Split(strings.TrimSpace(strings.TrimPrefix(text, "openapi:produces")), " ")
			op.Produces = append(op.Produces, parts...)
		} else if strings.HasPrefix(text, "openapi:body") {
			parts := strings.Split(strings.TrimSpace(strings.TrimPrefix(text, "openapi:body")), "---")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid openapi:body format: %s", name)
			}
			op.RequestBody.Name = strings.TrimSpace(parts[0])
			op.RequestBody.Description = strings.TrimSpace(parts[1])
		} else if strings.HasPrefix(text, "openapi:response") {
			res := &ResponseBody{}
			parts := strings.Split(strings.TrimSpace(strings.TrimPrefix(text, "openapi:response")), "---")
			if len(parts) > 2 {
				return nil, fmt.Errorf("invalid openapi:response format: %s", name)
			} else if len(parts) == 2 {
				res.Description = strings.TrimSpace(parts[1])
			}
			parts = strings.Split(strings.TrimSpace(parts[0]), " ")
			switch len(parts) {
			case 1:
				res.Code = parts[0]
			case 2:
				res.Code = parts[0]
				res.Name = parts[1]
			default:
				return nil, fmt.Errorf("invalid openapi:response format: %s", name)
			}
			op.Responses = append(op.Responses, res)
		} else if strings.HasPrefix(text, "openapi:param") {
			p := &Parameter{}
			parts := strings.Split(strings.TrimSpace(strings.TrimPrefix(text, "openapi:param")), "---")

			if len(parts) == 2 {
				p.Description = strings.TrimSpace(parts[1])
			}

			parts = strings.Split(strings.TrimSpace(parts[0]), " ")

			if len(parts) != 4 {
				return nil, fmt.Errorf("invalid openapi:param format: %s", name)
			}
			p.Name = parts[0]
			p.In = parts[1]
			p.Type = parts[2]
			p.Required = parts[3]
			op.Parameters = append(op.Parameters, p)
		}
	}

	if !isValidOperation {
		return nil, fmt.Errorf("%s does not contain openapi:operation shema", name)
	}

	return op, nil
}
