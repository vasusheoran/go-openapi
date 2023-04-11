package scan

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"go/ast"
	"strings"
)

func (p *Parser) ParseInterfaceType(ts *ast.TypeSpec) {
	interfaceType := ts.Type.(*ast.InterfaceType)
	if interfaceType.Methods == nil || len(interfaceType.Methods.List) == 0 {
		// If the interface has no methods, there's nothing to do.
		return
	}

	// Parse the comments for the interface.
	c, ok := p.comments[ts.Name.Name]
	if !ok {
		fmt.Printf("no comments found for method %s\n", ts.Name.Name)
	}
	ic := getInterfaceComments(c)

	for i, m := range interfaceType.Methods.List {
		// Parse the comments for the method.
		c, ok = p.comments[m.Names[0].Name]
		if !ok {
			fmt.Printf("no comments found for method %s\n", m.Names[i].Name)
			continue
		}
		mc := p.getMethodComments(c)
		if mc == nil {
			fmt.Printf("no comments found for method %s, continue to parse next method", m.Names[i].Name)
			continue
		}

		// Create the operation for the method and add it to the appropriate path in the OpenAPI spec.
		op := &openapi3.Operation{}
		if ic.Summary != "" {
			op.Summary = ic.Summary
		} else {
			op.Summary = strings.ReplaceAll(mc.Summary, "\n", " ")
		}
		if ic.Description != "" {
			op.Description = ic.Description
		} else {
			op.Description = strings.ReplaceAll(mc.Description, "\n", " ")
		}
		op.OperationID = ic.ID
		if ic.Tags != nil {
			op.Tags = ic.Tags
		} else {
			op.Tags = mc.Tags
		}

		// TODO: Set the operation's request and response parameters.
		// op.RequestBody
		// op.Responses

		op.RequestBody = p.ParseRequestType(mc)

		successResponse, err := p.ParseResponseType(mc.Success)
		if err != nil {
			fmt.Println(err)
			continue
		}

		failureResponse, err := p.ParseResponseType(mc.Failure)
		if err != nil {
			fmt.Println(err)
			continue
		}

		op.Responses = make(openapi3.Responses)

		if successResponse != nil {
			op.Responses["200"] = successResponse
		}

		if failureResponse != nil {
			op.Responses["default"] = failureResponse
		}

		// Set the operation tags.
		if len(mc.Tags) > 0 {
			op.Tags = mc.Tags
		}

		// Get or create the path item for the method's path.
		path := strings.Join([]string{ic.Path, mc.Path}, "")
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
