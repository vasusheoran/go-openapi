package scan

import (
	"go/ast"
	"strings"
)

type structComment struct {
	Description string
	OperationID string
}

type fieldComment struct {
	Description string
	Example     string
	Deprecated  bool
	Nullable    bool
	Format      string
	Default     string
	Enum        []interface{}
	//Type        string // Extracted from field schema
}

type tagOptions map[string]bool

func (opts tagOptions) Contains(key string) bool {
	_, ok := opts[key]
	return ok
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
