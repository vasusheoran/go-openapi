package main

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// SampleResponse1 This is a sample response struct comment
type SampleResponse1 struct {
	// This is a sample field comment
	// openapi:description Sample3 field description
	// openapi:example "sample3"
	// openapi:deprecated true
	// openapi:nullable true
	// openapi:format email
	// openapi:default "default_value"
	Field4 string `json:"field4"`
}

// SampleRequest2 This is a sample request struct comment
type SampleRequest2 struct {
	// This is a sample field comment
	// openapi:description Sample4 field description
	// openapi:example "sample4"
	// openapi:deprecated true
	// openapi:nullable true
	// openapi:format date-time
	// openapi:default "2023-04-12T14:01:00Z"
	Field5 string `json:"field5"`
	// This is a sample field comment
	// openapi:description Sample5 field description
	// openapi:minimum 1
	// openapi:maximum 100
	// openapi:exclusiveMaximum true
	Field6          int             `json:"field6"`
	SampleResponse1 SampleResponse1 `json:"sample_response_1"`
}

// SampleRequest1 This is a sample req struct comment
type SampleRequest1 struct {
	// This is a sample field comment
	// openapi:description Sample3 field description
	// openapi:example "sample3"
	Field3 string `json:"field3"`
}

// SampleInterface This is a sample interface comment
// openapi:server base_url
type SampleInterface interface {
	// SampleMethod1 This is a sample method 1 comment
	// openapi:summary Sample method 1 summary
	// openapi:description Sample method 1 description
	// openapi:tags sample1
	// openapi:id sampleMethod1
	// openapi:path /sample/method/{param2}
	// openapi:method GET
	// openapi:body SampleRequest1
	// openapi:param param1 query string true "Sample parameter 1"
	// openapi:param param2 path string true "Sample parameter 2"
	// openapi:param param3 header string false "Sample parameter 3"
	// openapi:success 200 SampleResponse1
	// openapi:failure 400 ErrorResponse
	SampleMethod1(param1 string, param2 string, param3 string) (*SampleResponse1, error)
	// SampleMethod2 This is a sample method 2 comment
	// openapi:summary Sample method 2 summary
	// openapi:description Sample method 2 description
	// openapi:tags sample2
	// openapi:id sampleMethod2
	// openapi:path /sample/method/{param5}
	// openapi:method POST
	// openapi:param param4 query string true "Sample parameter 1"
	// openapi:param param5 path string true "Sample parameter 2"
	// openapi:param param6 header string false "Sample parameter 3"
	// openapi:success 200 SampleResponse2
	// openapi:failure 400 ErrorResponse
	SampleMethod2(param4 string, param5 string, param6 string) (*SampleResponse2, error)
}

// SampleResponse2 This is a sample response struct comment
type SampleResponse2 struct {
	// This is a sample field comment
	// openapi:description Sample2 field description
	// openapi:example "sample2"
	Field2 string `json:"field2"`
}

// ErrorResponse This is a sample error response struct comment
type ErrorResponse struct {
	// This is a sample field comment for the error response
	// openapi:description Error message
	Message string `json:"message"`
}

func main() {}

// ParseInterfaceComments This function parses the comments for an interface and updates an OpenAPI spec
func ParseInterfaceComments(iface interface{}, spec *openapi3.T) error {
	// TODO: Implement parsing of interface comments to update OpenAPI specs.
	return nil
}
