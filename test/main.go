package main

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type FileData struct {
	Path    string
	Package string
	Types   map[string]*openapi3.SchemaRef
	Vars    map[string]interface{}
	Consts  map[string]interface{}
}

func main() {
	// Create an empty schema
	schema := &openapi3.Schema{
		Type:       "object",
		Properties: map[string]*openapi3.SchemaRef{},
	}
	// Traverse the directory recursively
	fdm := make(map[string]*FileData)
	_, err := traverseDir("../scan/testdata/pets", schema, fdm)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("OpenAPI 3.0 schema: %v\n", fdm)
}

func traverseDir(path string, schema *openapi3.Schema, fileDataMap map[string]*FileData) (map[string]*FileData, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if f.IsDir() {
			_, err = traverseDir(filepath.Join(path, f.Name()), schema, fileDataMap)
			if err != nil {
				return nil, err
			}
		} else {

			// Parse only .go files
			if filepath.Ext(f.Name()) != ".go" {
				continue
			}

			// Parse the file
			fileData, ok := fileDataMap[f.Name()]
			if !ok {
				fileData, err = parseFile(filepath.Join(path, f.Name()))
				if err != nil {
					return nil, err
				}
				fileDataMap[f.Name()] = fileData
			}
			// Add types to the schema
			for typeName, typeSchemaRef := range fileData.Types {
				if _, ok := schema.Properties[typeName]; !ok {
					schema.Properties[typeName] = typeSchemaRef
				}
			}
			// Add vars to the schema
			for varName, varValue := range fileData.Vars {
				schema.Properties[varName] = createSchemaFromValue(varValue)
			}
			// Add consts to the schema
			for constName, constValue := range fileData.Consts {
				schema.Properties[constName] = createSchemaFromValue(constValue)
			}
		}
	}
	return fileDataMap, err
}

func parseFile(path string) (*FileData, error) {
	// Parse the file
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	// Create a new FileData struct
	fileData := &FileData{
		Path:    path,
		Package: astFile.Name.Name,
		Types:   make(map[string]*openapi3.SchemaRef),
		Vars:    make(map[string]interface{}),
		Consts:  make(map[string]interface{}),
	}

	// Extract types, vars, and consts from the file
	for _, decl := range astFile.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			switch d.Tok {
			case token.TYPE:
				// Extract struct types

				for _, spec := range d.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}
					typeName := typeSpec.Name.Name
					typeSchema := createSchemaFromStruct(structType)
					fileData.Types[typeName] = &openapi3.SchemaRef{
						Value: typeSchema,
					}
				}
			case token.VAR:
				// Extract variables
				for _, spec := range d.Specs {
					valueSpec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for i, name := range valueSpec.Names {
						varName := name.Name
						varValue := valueSpec.Values[i]
						fileData.Vars[varName] = evalExpr(varValue)
					}
				}
			case token.CONST:
				// Extract constants
				for _, spec := range d.Specs {
					valueSpec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for i, name := range valueSpec.Names {
						constName := name.Name
						constValue := valueSpec.Values[i]
						fileData.Consts[constName] = evalExpr(constValue)
					}
				}
			}
		}
	}

	return fileData, nil
}

func createSchemaFromStruct(structType *ast.StructType) *openapi3.Schema {
	// Create a new schema for the struct type
	schema := &openapi3.Schema{
		Type:       "object",
		Properties: map[string]*openapi3.SchemaRef{},
	}

	// Add fields to the schema
	for _, field := range structType.Fields.List {
		fieldName := ""
		if field.Names != nil {
			fieldName = field.Names[0].Name
		} else if len(field.Names) == 0 && field.Type != nil {
			// Handle embedded fields
			if ident, ok := field.Type.(*ast.Ident); ok {
				fieldName = ident.Name
			}
		}
		if fieldName == "" {
			continue
		}
		fieldSchema := createSchemaFromExpr(field.Type)
		if field.Tag != nil {
			tags, err := parseTags(field.Tag.Value)
			if err == nil {
				if tags["json"] != "" {
					fieldName = tags["json"]
				}
			}
		}
		schema.Properties[fieldName] = fieldSchema
	}

	return schema
}

func createSchemaFromExpr(expr ast.Expr) *openapi3.SchemaRef {
	switch t := expr.(type) {
	case *ast.Ident:
		// Handle built-in types
		switch t.Name {
		case "bool":
			return openapi3.NewSchemaRef("", openapi3.NewBoolSchema())
		case "int", "int8", "int16", "int32", "int64":
			return openapi3.NewSchemaRef("", openapi3.NewIntegerSchema().WithFormat(t.Name))
		case "uint", "uint8", "uint16", "uint32", "uint64":
			return openapi3.NewSchemaRef("", openapi3.NewIntegerSchema().WithFormat(t.Name))
		case "float32", "float64":
			return openapi3.NewSchemaRef("", openapi3.NewFloat64Schema().WithFormat(t.Name))
		case "string":
			return openapi3.NewSchemaRef("", openapi3.NewStringSchema())
		case "error":
			return openapi3.NewSchemaRef("", openapi3.NewStringSchema())
		default:
			return openapi3.NewSchemaRef("", openapi3.NewStringSchema())
		}
	case *ast.ArrayType:
		itemSchema := createSchemaFromExpr(t.Elt)
		return itemSchema
	case *ast.MapType:
		valueSchema := createSchemaFromExpr(t.Value)
		return valueSchema
	case *ast.StarExpr:
		return createSchemaFromExpr(t.X)
	case *ast.SelectorExpr:
		return openapi3.NewSchemaRef(fmt.Sprintf("#/components/schemas/%s", t.Sel.Name), nil)
	default:
		return openapi3.NewSchemaRef("", openapi3.NewStringSchema())
	}
}

func evalExpr(expr ast.Expr) interface{} {
	switch t := expr.(type) {
	case *ast.BasicLit:
		switch t.Kind {
		case token.INT:
			val, _ := strconv.Atoi(t.Value)
			return val
		case token.FLOAT:
			val, _ := strconv.ParseFloat(t.Value, 64)
			return val
		case token.STRING:
			return t.Value[1 : len(t.Value)-1]
		default:
			return nil
		}
	case *ast.CompositeLit:
		switch t.Type.(type) {
		case *ast.ArrayType:
			arr := make([]interface{}, len(t.Elts))
			for i, elt := range t.Elts {
				arr[i] = evalExpr(elt)
			}
			return arr
		case *ast.MapType:
			m := make(map[string]interface{})
			for _, elt := range t.Elts {
				kv := elt.(*ast.KeyValueExpr)
				key := kv.Key.(*ast.BasicLit).Value
				m[key[1:len(key)-1]] = evalExpr(kv.Value)
			}
			return m
		default:
			return nil
		}
	default:
		return nil
	}
}

func parseTags(tag string) (map[string]string, error) {
	tags := make(map[string]string)
	tagList := strings.Split(tag, " ")
	for _, tag := range tagList {
		if strings.HasPrefix(tag, "`") {
			continue
		}
		tagParts := strings.Split(tag, ":")
		if len(tagParts) != 2 {
			return nil, fmt.Errorf("invalid tag: %s", tag)
		}
		tags[tagParts[0]] = tagParts[1]
	}
	return tags, nil
}
func createSchemaFromValue(v interface{}) *openapi3.SchemaRef {
	switch t := v.(type) {
	case bool:
		return openapi3.NewSchemaRef("", openapi3.NewBoolSchema())
	case int:
		return openapi3.NewSchemaRef("", openapi3.NewIntegerSchema())
	case int64:
		return openapi3.NewSchemaRef("", openapi3.NewInt64Schema())
	case float64:
		return openapi3.NewSchemaRef("", openapi3.NewFloat64Schema())
	case string:
		return openapi3.NewSchemaRef("", openapi3.NewStringSchema())
	case []interface{}:
		itemSchema := createSchemaFromValue(t[0])
		return openapi3.NewSchemaRef(itemSchema.Ref, openapi3.NewArraySchema().WithItems(itemSchema.Value))
	case map[string]interface{}:
		props := make(map[string]*openapi3.Schema)
		for k, v := range t {
			props[k] = createSchemaFromValue(v).Value
		}
		return openapi3.NewSchemaRef("", openapi3.NewObjectSchema().WithProperties(props))
	default:
		return openapi3.NewSchemaRef("", openapi3.NewStringSchema())
	}
}
