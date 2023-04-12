package errors

// ErrorResponse This is a sample error response struct comment`
type ErrorResponse struct {
	// This is a sample field comment
	// openapi:description Error message
	// openapi:example "404 not found"
	// openapi:deprecated true
	// openapi:nullable true
	// openapi:format text
	// openapi:default "404 not found"
	Msg string `json:"msg"`
}
