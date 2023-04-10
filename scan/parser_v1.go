package openapi

//
//import (
//	"bufio"
//	"encoding/json"
//	"errors"
//	"fmt"
//	"github.com/getkin/kin-openapi/openapi3"
//	"github.com/gorilla/mux"
//	"go/ast"
//	"go/parser"
//	"go/token"
//	"go/types"
//	"reflect"
//	"strings"
//)
//
//type parserClient struct {
//	info          *openapi3.Info
//	schemas       map[string]*openapi3.Schema
//	paths         map[string]*openapi3.PathItem
//	security      openapi3.SecurityRequirements
//	packages map[string]*ast.Package
//}
//
//func NewParser(info *openapi3.Info) *parserClient {
//	return &parserClient{
//		info:     info,
//		schemas:  make(map[string]*openapi3.Schema),
//		paths:    make(map[string]*openapi3.PathItem),
//		security: nil,
//	}
//}
//func (p *parserClient) parse(packagePath string) error {
//	fs := token.NewFileSet()
//	packages, err := parser.parseDir(fs, packagePath, nil, parser.parseComments)
//	if err != nil {
//		return err
//	}
//
//	for _, pkg := range packages {
//		for _, file := range pkg.Files {
//			if err := p.parseFile(file); err != nil {
//				return err
//			}
//		}
//	}
//
//	if len(p.paths) == 0 {
//		return errors.New("no API Path found")
//	}
//
//	return nil
//}
//
//func (p *parserClient) parseFile(file *ast.File) error {
//	for _, decl := range file.Decls {
//		fn, ok := decl.(*ast.FuncDecl)
//		if !ok {
//			continue
//		}
//
//		if err := p.parseEndpoint(file.Name.Name, fn); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (p *parserClient) parseEndpoint(pkgName string, fn *ast.FuncDecl) error {
//	// Extract the HTTP Method and Path from the function annotations
//	Method, Path, err := p.extractEndpointAnnotations(fn)
//	if err != nil {
//		return err
//	}
//
//	// Parse the input and output types for the Path
//	inputSchema, outputSchema, err := p.extractEndpointTypes(pkgName, fn)
//	if err != nil {
//		return err
//	}
//
//	// Create a new OpenAPI Path item for this Path
//	pathItem := &openapi3.PathItem{}
//
//	// Set the corresponding Path for the HTTP Method
//	switch Method {
//	case "GET":
//		pathItem.Get = &openapi3.Operation{
//			RequestBody: &openapi3.RequestBodyRef{
//				Value: &openapi3.RequestBody{
//					Content: openapi3.NewContentWithJSONSchema(inputSchema.Value),
//				},
//			},
//			Responses: openapi3.Responses{
//				"200": &openapi3.ResponseRef{
//					Value: &openapi3.Response{
//						Content: openapi3.NewContentWithJSONSchema(outputSchema.Value),
//					},
//				},
//			},
//		}
//	case "POST":
//		pathItem.Post = &openapi3.Operation{}
//	case "PUT":
//		pathItem.Put = &openapi3.Operation{}
//	case "PATCH":
//		pathItem.Patch = &openapi3.Operation{}
//	case "DELETE":
//		pathItem.Delete = &openapi3.Operation{}
//	case "OPTIONS":
//		pathItem.Options = &openapi3.Operation{}
//	case "HEAD":
//		pathItem.Head = &openapi3.Operation{}
//	default:
//		return fmt.Errorf("unsupported HTTP Method '%s'", Method)
//	}
//
//	// Set the Summary and Description for the Path
//	Summary, Description, err := p.extractEndpointDoc(fn)
//	if err != nil {
//		return err
//	}
//	if Summary != "" {
//		pathItem.Summary = Summary
//	}
//	if Description != "" {
//		pathItem.Description = Description
//	}
//
//	// Add the Path item to the OpenAPI paths map
//	p.paths[Path] = pathItem
//
//	return nil
//}
//
//func (p *parserClient) getPackage(pkgPath string) (*ast.Package, error) {
//	if pkg, ok := p.packages[pkgPath]; ok {
//		return pkg, nil
//	}
//
//	fset := token.NewFileSet()
//	pkgs, err := parser.parseDir(fset, pkgPath, nil, parser.parseComments)
//	if err != nil {
//		return nil, err
//	}
//
//	for _, pkg := range pkgs {
//		p.packages[pkgPath] = pkg
//		return pkg, nil
//	}
//
//	return nil, fmt.Errorf("no package found in directory: %s", pkgPath)
//}
//
//func (p *parserClient) extractEndpointTypes(pkgName string, fn *ast.FuncDecl) (*openapi3.SchemaRef, *openapi3.SchemaRef, error) {
//	var inputSchema *openapi3.SchemaRef
//	var outputSchema *openapi3.SchemaRef
//
//	for _, param := range fn.Type.Params.List {
//		schemaType, err := p.extractParamType(pkgName, param)
//		if err != nil {
//			return nil, nil, err
//		}
//		if inputSchema == nil {
//			inputSchema = schemaType
//		} else {
//			inputSchema = &openapi3.SchemaRef{
//				Value: &openapi3.Schema{
//					OneOf: []*openapi3.SchemaRef{
//						inputSchema,
//						schemaType,
//					},
//				},
//			}
//		}
//	}
//
//	results := fn.Type.Results
//	if results != nil && len(results.List) > 0 {
//		returnType := results.List[0]
//		schemaType, err := p.extractType(pkgName, returnType.Type)
//		if err != nil {
//			return nil, nil, err
//		}
//		outputSchema.Value.Type = schemaType
//	}
//
//	return inputSchema, outputSchema, nil
//}
//
//func (p *parserClient) extractType(pkgName string, expr ast.Expr) (schemaType string, err error) {
//	switch t := expr.(type) {
//	case *ast.Ident:
//		schemaType = p.getTypeName(t)
//	case *ast.ArrayType:
//		schemaType, err = p.extractArrayType(pkgName, t)
//	case *ast.MapType:
//		schemaType, err = p.extractMapType(pkgName, t.Map)
//	case *ast.StarExpr:
//		schemaType, err = p.extractType(pkgName, t.X)
//	case *ast.SelectorExpr:
//		schemaType, err = p.extractSelectorType(pkgName, t)
//	default:
//		err = fmt.Errorf("unsupported return type: %s", expr)
//	}
//	return schemaType, err
//}
//
//func (p *parserClient) extractSelectorType(name string, t *ast.SelectorExpr) (string, error) {
//	ident, ok := t.X.(*ast.Ident)
//	if !ok {
//		return "", fmt.Errorf("unsupported expression type: %T", t.X)
//	}
//
//	pkgName := ident.Name
//	if ident.Obj != nil {
//		pkgName = ident.Obj.Name
//	}
//
//	typeName := t.Sel.Name
//	return p.extractNamedType(pkgName, typeName)
//}
//func (p *parserClient) extractNamedType(pkgName, typeName string) (string, error) {
//	// Search for the package among already processed ones.
//	if pkg, err := p.getPackage(pkgName); err != nil {
//
//		pkg.Scope
//		if t, ok := pkg.Types[typeName]; ok {
//			return p.extractType(pkgName, t.Type)
//		}
//	}
//
//	// If the package has not been processed yet, parse it.
//	fset := token.NewFileSet()
//	pkg, err := parser.parseDir(p.fsys, pkgName, nil, parser.parseComments)
//	if err != nil {
//		return "", err
//	}
//
//	// Cache the parsed package.
//	p.packages[pkgName] = pkg
//
//	// Extract the named type from the package.
//	var namedType *types.Named
//	for _, file := range pkg {
//		for _, decl := range file.Scope.Objects {
//			if typ, ok := decl.Decl.(*ast.TypeSpec); ok {
//				if named, ok := typ.Type.(*ast.Ident); ok && named.Name == typeName {
//					namedType, _ = p.info.TypeOf(named).(*types.Named)
//					break
//				}
//			}
//		}
//	}
//
//	// If the named type was not found in the package, return an error.
//	if namedType == nil {
//		return "", fmt.Errorf("could not find named type %s in package %s", typeName, pkgName)
//	}
//
//	return p.extractType(pkgName, namedType.Underlying())
//}
//
////func (p *parserClient) extractSelectorType(pkgName string, t *types.Named) (*openapi3.SchemaRef, error) {
////	schema := &openapi3.SchemaRef{}
////
////	// Check if the named type is a selector expression
////	selector, ok := t.Underlying().(*types.Named).Obj().Type().(*types.Named)
////	if !ok {
////		return nil, nil
////	}
////
////	// Extract the selector type and create a reference to it
////	selectorType, err := p.extractType(pkgName, selector)
////	if err != nil {
////		return nil, err
////	}
////
////	schema.Ref = fmt.Sprintf("#/components/schemas/%s", selector.Obj().Name())
////	schema.Value = selectorType.Value
////
////	return schema, nil
////}
//
//// extractMapType extracts the schema type of a map variable
//func (p *parserClient) extractMapType(pkgName string, t *types.Map) (schemaType string, err error) {
//	valueType := t.Elem()
//	valueSchemaType, err := p.extractType(pkgName, t)
//	if err != nil {
//		return "", err
//	}
//
//	return fmt.Sprintf("object{additionalProperties:%s}", valueSchemaType), nil
//}
//
//func (p *parserClient) extractArrayType(pkgName string, t ast.Expr) (schemaType string, err error) {
//	// Check if the type is an array
//	arrayType, ok := t.(*ast.ArrayType)
//	if !ok {
//		return "", fmt.Errorf("not an array type")
//	}
//
//	// Check if the array is a slice of bytes
//	if _, ok := arrayType.Elt.(*ast.Ident); ok && arrayType.Elt.(*ast.Ident).Name == "byte" {
//		return "string", nil
//	}
//
//	// Check if the array is a slice
//	if sliceType, ok := arrayType.Elt.(*ast.SliceExpr); ok {
//		sliceSchemaType, err := p.extractArrayType(pkgName, sliceType)
//		if err != nil {
//			return "", err
//		}
//		return "array[" + sliceSchemaType + "]", nil
//	}
//
//	// Extract the type of the array
//	schemaType, err = p.extractType(pkgName, arrayType.Elt)
//	if err != nil {
//		return "", err
//	}
//
//	return "array[" + schemaType + "]", nil
//}
//
//func (p *parserClient) getTypeName(typeExpr ast.Expr) string {
//	switch expr := typeExpr.(type) {
//	case *ast.Ident:
//		return expr.Name
//	case *ast.SelectorExpr:
//		return fmt.Sprintf("%s.%s", p.getTypeName(expr.X), expr.Sel.Name)
//	case *ast.StarExpr:
//		return "*" + p.getTypeName(expr.X)
//	case *ast.ArrayType:
//		return "[]" + p.getTypeName(expr.Elt)
//	case *ast.MapType:
//		keyType := p.getTypeName(expr.Key)
//		valueType := p.getTypeName(expr.Value)
//		return fmt.Sprintf("map[%s]%s", keyType, valueType)
//	default:
//		return "unknown"
//	}
//}
//
//func (p *parserClient) extractParamType(pkgName string, param *ast.Field) (*openapi3.SchemaRef, error) {
//	t, ok := param.Type.(*ast.Ident)
//	if !ok {
//		return nil, fmt.Errorf("unsupported parameter type %s", param.Type)
//	}
//	var schemaType string
//	switch t.Name {
//	case "string":
//		schemaType = openapi3.TypeString
//	case "bool":
//		schemaType = openapi3.TypeBoolean
//	case "int":
//		schemaType = openapi3.TypeInteger
//	case "float32", "float64":
//		schemaType = openapi3.TypeNumber
//	default:
//		return nil, fmt.Errorf("unsupported parameter type %s", t.Name)
//	}
//
//	return &openapi3.SchemaRef{
//		Value: &openapi3.Schema{
//			Type: schemaType,
//		},
//	}, nil
//}
//
//func (p *parserClient) extractEndpointDoc(fn *ast.FuncDecl) (Summary, Description string, err error) {
//	doc := fn.Doc
//	if doc == nil {
//		return "", "", nil
//	}
//
//	Summary, Description, err = parseDoc(doc.Text())
//	if err != nil {
//		return "", "", fmt.Errorf("error parsing Path doc for function %s: %w", fn.Name.Name, err)
//	}
//
//	return Summary, Description, nil
//}
//
//func parseDoc(doc string) (Summary, Description string, err error) {
//	scanner := bufio.NewScanner(strings.NewReader(doc))
//	var lines []string
//	for scanner.Scan() {
//		lines = append(lines, scanner.Text())
//	}
//	if err := scanner.Err(); err != nil {
//		return "", "", fmt.Errorf("error parsing Path doc: %w", err)
//	}
//
//	if len(lines) == 0 {
//		return "", "", nil
//	}
//
//	// Summary is the first line of the doc
//	Summary = lines[0]
//
//	// Description is everything else except for empty lines
//	var descLines []string
//	for _, line := range lines[1:] {
//		trimmed := strings.TrimSpace(line)
//		if trimmed != "" {
//			descLines = append(descLines, trimmed)
//		}
//	}
//	Description = strings.Join(descLines, " ")
//
//	return Summary, Description, nil
//}
//
//func (p *parserClient) extractEndpointAnnotations(fn *ast.FuncDecl) (Method, Path string, err error) {
//	// Extract the Method and Path annotations from the function declaration
//	for _, c := range fn.Doc.List {
//		if strings.HasPrefix(c.Text, "// @") {
//			parts := strings.Split(c.Text[3:], " ")
//			switch parts[0] {
//			case "Method":
//				Method = strings.TrimSpace(parts[1])
//			case "Path":
//				Path = strings.TrimSpace(parts[1])
//			}
//		}
//	}
//
//	// Validate the extracted Method and Path values
//	if Method == "" {
//		err = fmt.Errorf("missing 'Method' annotation in function '%s'", fn.Name.Name)
//	}
//	if Path == "" {
//		err = fmt.Errorf("missing 'Path' annotation in function '%s'", fn.Name.Name)
//	}
//
//	return Method, Path, err
//}
//
//func parseParameters(Path string, Method string, handlerFuncType reflect.Type) ([]*openapi3.ParameterRef, error) {
//	params := mux.Vars(Path)
//	paramRefs := make([]*openapi3.ParameterRef, len(params))
//
//	for i, param := range params {
//		schema := openapi3.NewSchemaRef("", openapi3.NewStringSchema())
//		paramRef := &openapi3.ParameterRef{
//			Ref: "",
//			Value: &openapi3.Parameter{
//				Name:   param,
//				In:     "Path",
//				Schema: schema,
//			},
//		}
//		paramRefs[i] = paramRef
//	}
//
//	requestBodySchemaRef := openapi3.NewSchemaRef("", openapi3.NewObjectSchema())
//
//	requestContentType := getAnnotationValue(handlerFuncType, "Accept")
//	requestBodyParamRef := &openapi3.ParameterRef{
//		Ref: "",
//		Value: &openapi3.Parameter{
//			Name:        "body",
//			In:          "body",
//			Description: "The request body",
//			Required:    true,
//			Schema:      requestBodySchemaRef,
//		},
//	}
//
//	params := append(paramRefs, requestBodyParamRef)
//	return params, nil
//}
//func parseExamples(handlerFuncType reflect.Type) (*openapi3.Example, *openapi3.Example, error) {
//	requestExample := &openapi3.Example{}
//	responseExample := &openapi3.Example{}
//
//	requestExampleAnnotation := getAnnotation(handlerFuncType, "RequestExample")
//	if requestExampleAnnotation != nil {
//		err := json.Unmarshal([]byte(requestExampleAnnotation.Value), &requestExample.Value)
//		if err != nil {
//			return nil, nil, fmt.Errorf("could not parse request example: %v", err)
//		}
//	}
//
//	responseExampleAnnotation := getAnnotation(handlerFuncType, "ResponseExample")
//	if responseExampleAnnotation != nil {
//		err := json.Unmarshal([]byte(responseExampleAnnotation.Value), &responseExample.Value)
//		if err != nil {
//			return nil, nil, fmt.Errorf("could not parse response example: %v", err)
//		}
//	}
//
//	return requestExample, responseExample, nil
//}
//
//func createEndpoint(Path string, Method string, Path *openapi3.Operation) (*Endpoint, error) {
//	Path := &Endpoint{
//		Path:   Path,
//		Method: strings.ToUpper(Method),
//		//Summary: Path.Summary,
//		//Description: Path.Description,
//		//OperationID: Path.OperationID,
//		Parameters: Path.Parameters,
//		//RequestBody: Path.RequestBody,
//		Responses: Path.Responses,
//	}
//
//	return Path, nil
//}

//func (p *parserClient) processPackage(pkgName string) error {
//	// If this package is already processed, return early
//	if p.processed[pkgName] {
//		return nil
//	}
//
//	p.processed[pkgName] = true
//
//	pkg, ok := p.packages[pkgName]
//	if !ok {
//		return fmt.Errorf("package not found: %s", pkgName)
//	}
//
//	for _, file := range pkg.Files {
//		for _, decl := range file.Decls {
//			fn, ok := decl.(*ast.FuncDecl)
//			if !ok {
//				continue
//			}
//
//			if !p.parseComments(fn, fn.Doc) {
//				continue
//			}
//
//			Summary, Description, err := p.extractEndpointDoc(fn)
//			if err != nil {
//				return err
//			}
//
//			endpointPath, Method, err := p.extractEndpointAnnotations(fn)
//			if err != nil {
//				return err
//			}
//
//			requestBodySchemaRef, responseSchemaRef, err := p.extractEndpointTypes(pkgName, fn)
//			if err != nil {
//				return err
//			}
//
//			pathItem := p.generatePathItem(Summary, Description, Method, &openapi3.Operation{
//				RequestBody: requestBodySchemaRef,
//				Responses:   *responseSchemaRef,
//			})
//
//			if _, ok := p.doc.Paths[endpointPath]; !ok {
//				p.doc.Paths[endpointPath] = &openapi3.PathItem{}
//			}
//
//			p.doc.Paths[endpointPath] = pathItem
//		}
//	}
//
//	for _, importSpec := range pkg.Imports {
//		importPath := importSpec.Name[1 : len(importSpec.Name)-1]
//		if _, ok := p.packages[importPath]; !ok {
//			if err := p.processPackage(importPath); err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}
