package scan

import (
	"github.com/getkin/kin-openapi/openapi3"
	"go/ast"
	"path/filepath"
	"strconv"
	"strings"
)

func (p *Parser) ParseInterfaceType(ts *ast.TypeSpec) {
	interfaceType := ts.Type.(*ast.InterfaceType)
	if interfaceType.Methods == nil || len(interfaceType.Methods.List) == 0 {
		// If the interface has no methods, there's nothing to do.
		return
	}
	var ic *interfaceComments
	// Parse the comments for the interface and add tags
	c, ok := p.comments[ts.Name.Name]
	if ok {
		ic = getInterfaceComments(c)

		if len(ic.Name) == 0 {
			p.logger.Error("must have `openapi:name` annotation for %s", ts.Name.Name)
			return
		}

		tag := &openapi3.Tag{
			Name:        ic.Name,
			Description: ic.Description,
		}

		ed := openapi3.ExternalDocs{}
		if len(ic.ExternalDocs.URL) > 0 {
			ed.URL = ic.ExternalDocs.URL
		}
		if len(ic.ExternalDocs.Description) > 0 {
			ed.Description = ic.ExternalDocs.Description
		}

		if len(ic.ExternalDocs.URL) > 0 || len(ic.ExternalDocs.Description) > 0 {
			tag.ExternalDocs = &ed
		}

		p.spec.Tags = append(p.spec.Tags, tag)
	} else {
		p.logger.Debug("no comments found for %s", ts.Name.Name)
	}

	for _, field := range interfaceType.Methods.List {
		op := &openapi3.Operation{}

		key := filepath.Join(ts.Name.Name, field.Names[0].Name)

		// Parse the comments for the method.
		c, ok = p.comments[key]
		if !ok {
			p.logger.Warn("no comments found for %s", key)
			continue
		}
		mc := p.getMethodComments(c, key)
		if mc == nil {
			p.logger.Warn("failed to parse comments for %s", key)
			continue
		}

		if mc.Summary != "" {
			op.Summary = mc.Summary
		} else {
			op.Summary = strings.ReplaceAll(mc.Summary, "\n", " ")
		}
		if mc.Description != "" {
			op.Description = mc.Description
		} else {
			op.Description = strings.ReplaceAll(mc.Description, "\n", " ")
		}
		op.OperationID = mc.ID
		if mc.Tags != nil {
			op.Tags = mc.Tags
		} else {
			op.Tags = mc.Tags
		}

		// TODO: Set the operation's request and response parameters.
		// op.RequestBody
		// op.Responses

		op.RequestBody = getRequestBodyFromMethodComments(mc)

		op.Parameters = getParametersFromMethodComments(mc.Parameters)

		successResponse, err := getResponseFromMethodComments(mc.Success)
		if err != nil {
			p.logger.Warn("failed to parse error response for %s", key)
			continue
		}

		failureResponse, err := getResponseFromMethodComments(mc.Failure)
		if err != nil {
			p.logger.Warn("failed to parse error response for %s", key)
			continue
		}

		op.Responses = make(openapi3.Responses)

		if successResponse != nil {
			op.Responses[strconv.Itoa(mc.Success.Code)] = successResponse
		}

		if failureResponse != nil {
			op.Responses[strconv.Itoa(mc.Failure.Code)] = failureResponse
		}

		// Set the operation tags.
		if len(mc.Tags) > 0 {
			op.Tags = mc.Tags
		}

		// Get or create the path item for the method's path.
		path := strings.Join([]string{ic.BasePath, mc.Path}, "")
		pathItem := p.getPathItem(path)
		if p == nil {
			pathItem = &openapi3.PathItem{}
		}

		// Add the operation to the path item.
		switch strings.ToUpper(mc.Method) {
		case "GET":
			pathItem.Get = op
		case "PUT":
			pathItem.Put = op
		case "POST":
			pathItem.Post = op
		case "DELETE":
			pathItem.Delete = op
		case "OPTIONS":
			pathItem.Options = op
		case "HEAD":
			pathItem.Head = op
		case "PATCH":
			pathItem.Patch = op
		case "TRACE":
			pathItem.Trace = op
		default:
			// If the method name isn't recognized, skip it.
			p.logger.Warn("unrecognized method for %s: %s", key, mc.Method)
			continue
		}

		p.spec.Paths[path] = pathItem
	}
}

func (p *Parser) getPathItem(path string) *openapi3.PathItem {
	// If the path item does not exist, create it and add it to the paths map.
	if _, ok := p.paths[path]; !ok {
		p.paths[path] = &openapi3.PathItem{}
	}

	return p.paths[path]
}
