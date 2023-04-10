package openapi

import (
	"fmt"
	"go/ast"
	"strings"
)

func parseOpenAPITags(comment *ast.CommentGroup) (Endpoint, error) {
	ep := Endpoint{}

	if comment == nil {
		return ep, nil
	}

	prefix := "// @openapi:"
	for _, c := range comment.List {

		if !strings.HasPrefix(c.Text, prefix) {
			continue
		}

		parts := strings.SplitN(strings.TrimPrefix(c.Text, prefix), " ", 2)
		if len(parts) < 2 {
			continue
		}

		fmt.Println("parsing " + parts[0])

		switch parts[0] {
		case "path":
			ep.Path = strings.Trim(parts[1], " ")
		case "method":
			ep.Method = strings.ToUpper(strings.Trim(parts[1], " "))
		case "summary":
			ep.Summary = strings.Trim(parts[1], " ")
		case "description":
			ep.Description = strings.Trim(parts[1], " ")
		case "operationId":
			ep.OperationID = strings.Trim(parts[1], " ")
			//case "response":
			//	if len(parts) < 3 {
			//		continue
			//	}
			//	var r Response
			//	r.Code, _ = strconv.Atoi(strings.Trim(parts[1], " "))
			//	parts = strings.SplitN(strings.Trim(parts[2], " "), " ", 2)
			//	r.Type = strings.Trim(parts[0], " ")
			//	r.Description = strings.Trim(parts[1], " ")
			//	ep.Responses = append(ep.Responses, r)
			//case "security":
			//	parts := strings.Split(strings.Trim(parts[1], " "), " ")
			//	var security Security
			//	security.Name = parts[0]
			//	security.Scopes = make([]string, len(parts)-1)
			//	for i, s := range parts[1:] {
			//		security.Scopes[i] = s
			//	}
			//	ep.Security = append(ep.Security, security)
			//case "query":
			//	if len(parts) < 4 {
			//		continue
			//	}
			//	var qp QueryParam
			//	qp.Name = strings.Trim(parts[1], " ")
			//	qp.Type = strings.Trim(parts[2], " ")
			//	qp.Required, _ = strconv.ParseBool(strings.Trim(parts[3], " "))
			//	if len(parts) > 4 {
			//		qp.Description = strings.Trim(parts[4], " ")
			//	}
			//	ep.QueryParams = append(ep.QueryParams, qp)
			//case "pathParam":
			//	if len(parts) < 4 {
			//		continue
			//	}
			//	var pp PathParam
			//	pp.Name = strings.Trim(parts[1], " ")
			//	pp.Type = strings.Trim(parts[2], " ")
			//	pp.Description = strings.Trim(parts[3], " ")
			//	ep.PathParams = append(ep.PathParams, pp)
		}
	}
	return ep, nil
}
