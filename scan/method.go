package scan

import (
	"github.com/getkin/kin-openapi/openapi3"
)

type externalDocs struct {
	URL         string
	Description string
}

type interfaceComments struct {
	Name         string
	Description  string
	BasePath     string
	ExternalDocs externalDocs
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
	Body        string
}

type responseComments struct {
	Code        int
	Description string
	Schema      *openapi3.SchemaRef
}
