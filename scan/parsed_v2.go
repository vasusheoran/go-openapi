package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/imdario/mergo"
)

type Endpoint struct {
	Method            string
	Path              string
	Summary           string
	Description       string
	Name              string
	RequestType       string
	OperationID       string
	Tags              []string
	Parameters        openapi3.Parameters
	RequestBodyTypes  openapi3.RequestBodyRef
	ResponseBodyTypes openapi3.Responses
	Operation         openapi3.Operation
}

type parserClient struct {
	fs        *token.FileSet
	doc       *openapi3.T
	processed map[string]bool
	packages  map[string]*ast.Package
	endpoints map[string]Endpoint
}

func NewParser() *parserClient {
	return &parserClient{
		fs:        token.NewFileSet(),
		doc:       &openapi3.T{},
		processed: map[string]bool{},
		packages:  make(map[string]*ast.Package),
		endpoints: map[string]Endpoint{},
	}
}

func (p *parserClient) log(message string, keys ...string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf("[%s] %s\n", timestamp, message)
	fmt.Printf(msg, keys)
}

func (p *parserClient) GenerateOpenAPI(dirPath string) (*openapi3.T, error) {
	if err := p.parseDir(dirPath); err != nil {
		return nil, err
	}
	return p.doc, nil
}

func (p *parserClient) parseDir(dirPath string) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dirPath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		err = p.parsePackage(fset, pkg.Name, pkg, pkgs)
		if err != nil {
			return err
		}
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			err = p.parseFile(file)
			if err != nil {
				return err
			}
		}
	}

	// Update openAPI spec
	for key, ep := range p.endpoints {
		op := ep.Operation

		op.Summary = ep.Summary
		op.Description = ep.Description
		op.OperationID = key

		for _, param := range ep.Parameters {
			op.Parameters = append(op.Parameters, param)
		}

		pathItem := p.generatePathItem(ep.Summary, ep.Description, ep.Method, &openapi3.Operation{
			RequestBody: &ep.RequestBodyTypes,
			Responses:   ep.ResponseBodyTypes,
		})

		if p.doc.Paths == nil {
			p.doc.Paths = make(openapi3.Paths)
		}

		if _, ok := p.doc.Paths[ep.Path]; !ok {
			p.doc.Paths[ep.Path] = &openapi3.PathItem{}
		}

		p.doc.Paths[ep.Path] = pathItem
	}
	return nil
}

func (p *parserClient) parsePackage(fset *token.FileSet, pkgName string, pkg *ast.Package, pkgs map[string]*ast.Package) error {
	if _, ok := p.processed[pkgName]; ok {
		return nil
	}
	p.processed[pkgName] = true

	for _, imp := range pkg.Imports {
		if err := p.parsePackage(fset, imp.Name, pkgs[imp.Name], pkgs); err != nil {
			return fmt.Errorf("failed to parse package %s: %v", imp.Name, err)
		}
	}

	return nil
}

func (p *parserClient) parseFile(file *ast.File) error {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		comment := fn.Doc
		if comment == nil {
			continue
		}
		err := p.parseEndpoint(file.Name.Name, fn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *parserClient) parseOpenAPIComment(commentText string, operation openapi3.Operation) error {
	prefix := "// @openapi:"
	lines := strings.Split(commentText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			specStr := strings.TrimPrefix(line, prefix)
			specStr = strings.TrimSpace(specStr)
			specBytes := []byte(specStr)

			var openapiSpec map[string]interface{}
			err := json.Unmarshal(specBytes, &openapiSpec)
			if err != nil {
				return fmt.Errorf("failed to parse @openapi spec: %s", err.Error())
			}

			err = mergo.Merge(&operation, openapiSpec, mergo.WithOverride)
			if err != nil {
				return fmt.Errorf("failed to merge @openapi spec: %s", err.Error())
			}
		}
	}
	return nil
}

func (p *parserClient) parseEndpoint(pkgName string, fn *ast.FuncDecl) error {
	op := openapi3.Operation{}
	var ep1 Endpoint

	// Parse the comment for the Path
	if fn.Doc != nil {
		var err error
		ep1, err = parseOpenAPITags(fn.Doc)
		if err != nil {
			return err
		}
	}

	fmt.Println(ep1)

	operationID := p.extractOperationID(fn)

	// Set the Path's tags based on the package name
	//Path.Tags = append(Path.Tags, pkgName)

	// Set the Path's Summary based on the function name
	//Path.Summary = fn.Name.Name

	// Parse the function Parameters
	_, err := p.extractEndpointParams(fn, operationID)
	if err != nil {
		p.log("no Path params found %s", fn.Name.Name)
		//return nil, err
	}

	// Parse the function return type
	responses, err := p.extractEndpointResponse(fn)
	if err != nil {
		p.log("%s for %s", err.Error(), fn.Name.Name)
	}

	//if err == nil {
	//	Path.Responses = *response
	//}

	_, _, err = p.extractEndpointDoc(fn, operationID)
	if err != nil {
		return err
	}

	// extractEndpointTypes updates the parser struct to with reponse and request body
	_, _, err = p.extractEndpointTypes(pkgName, operationID, fn)
	if err != nil {
		p.log("no request or response body found types found %s", fn.Name.Name)
	}

	_, _, err = p.extractEndpointAnnotations(fn, operationID)
	if err != nil {
		// If current ast is a struct, then check if Path is already parsed.
		p.log("extractEndpointAnnotations returned error, check if Path already parsed")
		//return nil, err
	}

	endpointItem, ok := p.endpoints[operationID]
	if !ok {
		p.endpoints[operationID] = Endpoint{}
	} else {
		p.log("%s already exist for %s", operationID, fn.Name.Name)
	}

	if responses != nil {
		endpointItem.ResponseBodyTypes = *responses
	}
	endpointItem.Operation = op
	p.endpoints[operationID] = endpointItem
	return nil
}

func (p *parserClient) extractOperationID(fn *ast.FuncDecl) string {
	// Check if the function has a 'operationId' annotation
	if operationID, ok := p.getAnnotationValue(fn.Doc, "operationId"); ok {
		if _, ok := p.endpoints[operationID]; ok {
			p.log("opertionID %s exist for %s", operationID, fn.Name.Name)
		}
		return operationID
	}

	// If no 'operationId' annotation is found, use the function name as the 'operationID'
	return fn.Name.Name
}

func (p *parserClient) getAnnotationValue(doc *ast.CommentGroup, name string) (string, bool) {
	prefix := "// @openapi:" + name + " "
	if doc == nil {
		return "", false
	}

	for _, comment := range doc.List {
		if strings.HasPrefix(comment.Text, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(comment.Text, prefix)), true
		}
	}

	return "", false
}

func (p *parserClient) extractEndpointResponse(fn *ast.FuncDecl) (*openapi3.Responses, error) {
	// Check if the function has a return type
	if fn.Type.Results == nil {
		p.log("no return type for function '%s'", fn.Name.Name)
		return nil, errNoReturnType
	}

	// Check if the return type is a single value or a tuple
	numResults := fn.Type.Results.NumFields()
	if numResults == 0 {
		return nil, fmt.Errorf("no return type for function '%s'", fn.Name.Name)
	} else if numResults == 1 {
		// Single return value
		return p.extractSingleEndpointResponse(fn.Type.Results.List[0].Type)
	} else {
		// Tuple return value
		return p.extractTupleEndpointResponse(fn.Type.Results.List)
	}
}

func (p *parserClient) extractSingleEndpointResponse(respType ast.Expr) (*openapi3.Responses, error) {
	respRef, err := p.generateSchemaRef(respType)
	if err != nil {
		return nil, err
	}

	responses := &openapi3.Responses{
		strconv.Itoa(200): &openapi3.ResponseRef{
			Value: &openapi3.Response{
				//Description: respRef.,
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{
						Schema: respRef,
					},
				},
			},
		},
	}
	return responses, nil
}

func (p *parserClient) extractTupleEndpointResponse(respFields []*ast.Field) (*openapi3.Responses, error) {
	r := make(openapi3.Responses)

	for _, field := range respFields {
		if field == nil || field.Type == nil {
			return nil, fmt.Errorf("invalid response field")
		}

		respType, ok := field.Type.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("invalid response field type")
		}

		// Extract the status code from the field tag, if present
		statusCode := http.StatusOK
		if field.Tag != nil {
			tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
			codeStr := tag.Get("status")
			if codeStr != "" {
				code, err := strconv.Atoi(codeStr)
				if err != nil {
					return nil, fmt.Errorf("invalid status code: %s", codeStr)
				}
				statusCode = code
			}
		}

		// Extract the Description from the field comment, if present
		description := ""
		if field.Comment != nil {
			commentText := field.Comment.Text()
			if commentText != "" {
				description = strings.TrimSpace(commentText)
			}
		}

		// Generate the response schema reference
		respRef, err := p.generateSchemaRef(respType)
		if err != nil {
			return nil, err
		}

		response := &openapi3.Response{
			Description: &description,
			Content:     openapi3.NewContentWithJSONSchemaRef(respRef),
		}

		response.WithDescription(description)

		r[strconv.Itoa(statusCode)] = &openapi3.ResponseRef{Value: response}
	}
	return &r, nil
}

func (p *parserClient) extractEndpointParams(fn *ast.FuncDecl, operationID string) ([]*openapi3.ParameterRef, error) {
	p.log("extractEndpointParams: %s", fn.Name.Name)
	var params []*openapi3.ParameterRef
	for _, field := range fn.Type.Params.List {
		for _, name := range field.Names {
			schemaRef, err := p.generateSchemaRef(field.Type)
			if err != nil {
				return nil, err
			}
			param := &openapi3.Parameter{
				Name: name.Name,
				In:   "Path",
				Schema: &openapi3.SchemaRef{
					Value: schemaRef.Value,
				},
			}
			params = append(params, &openapi3.ParameterRef{
				Value: param,
			})
		}
	}

	ep, ok := p.endpoints[operationID]
	if !ok {
		ep = Endpoint{}
	}

	if ep.Parameters != nil {
		p.log("Parameters found for %s", operationID)
	}

	ep.Parameters = params

	return params, nil
}

func (p *parserClient) parseComments(fn *ast.FuncDecl, doc *ast.CommentGroup) (parsed bool) {
	if doc == nil {
		return
	}
	commentLines := strings.Split(doc.Text(), "\n")
	for _, commentLine := range commentLines {
		if strings.HasPrefix(commentLine, "//") {
			commentLine = strings.TrimPrefix(commentLine, "//")
			commentLine = strings.TrimSpace(commentLine)
		}
		if strings.HasPrefix(commentLine, "@openapi") {
			return true
		}
	}
	return false
}

func (p *parserClient) getPackage(pkgPath string) (*ast.Package, error) {
	if pkg, ok := p.packages[pkgPath]; ok {
		return pkg, nil
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkgPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		p.packages[pkgPath] = pkg
		return pkg, nil
	}

	return nil, fmt.Errorf("no package found in directory: %s", pkgPath)
}

func (p *parserClient) extractEndpointAnnotations(fn *ast.FuncDecl, operationID string) (string, string, error) {
	var path, method string
	for _, comment := range fn.Doc.List {
		if comment.Text == "" {
			continue
		}
		if matches := regexp.MustCompile(`@Path\s+(.+)`).FindStringSubmatch(comment.Text); len(matches) > 0 {
			path = matches[1]
		}
		if matches := regexp.MustCompile(`@Method\s+(.+)`).FindStringSubmatch(comment.Text); len(matches) > 0 {
			method = matches[1]
		}
	}
	if path == "" || method == "" {
		return "", "", errors.New("Path and Method annotations are required")
	}

	ep, ok := p.endpoints[operationID]
	if !ok {
		p.endpoints[operationID] = Endpoint{}
	}

	if len(ep.Path) > 0 || len(ep.Method) > 0 {
		p.log("Endpoint annotations found for %s", operationID)
	}

	ep.Path = path
	ep.Method = method

	p.endpoints[operationID] = ep
	return path, method, nil
}

// extractEndpointDoc extracts the Summary and Description annotations from the function declaration.
func (p *parserClient) extractEndpointDoc(fn *ast.FuncDecl, operationID string) (string, string, error) {
	var summary, description string

	// Retrieve the first comment group associated with the function.
	cg := fn.Doc
	if cg == nil {
		return summary, description, nil
	}

	// Extract the Summary and Description from the comment group.
	lines := strings.Split(cg.Text(), "\n")
	if len(lines) > 0 {
		summary = strings.TrimSpace(lines[0])
	}
	if len(lines) > 1 {
		description = strings.TrimSpace(strings.Join(lines[1:], "\n"))
	}

	ep, ok := p.endpoints[operationID]
	if !ok {
		p.endpoints = map[string]Endpoint{}
	}

	if len(ep.Summary) > 0 || len(ep.Description) > 0 {
		p.log("Endpoint docs found for %s", operationID)
	}

	ep.Summary = summary
	ep.Description = description

	p.endpoints[operationID] = ep
	return summary, description, nil
}

func (p *parserClient) extractEndpointTypes(pkgName string, operationID string, fn *ast.FuncDecl) (*openapi3.RequestBodyRef, *openapi3.Responses, error) {
	requestBodyRef := &openapi3.RequestBodyRef{}
	responses := &openapi3.Responses{}
	var err error

	if fn.Type.Params != nil && len(fn.Type.Params.List) > 0 {
		params := fn.Type.Params.List
		requestBody := openapi3.NewRequestBody()
		requestBody.Required = true
		requestBody.Content = make(map[string]*openapi3.MediaType)

		for _, param := range params {
			//paramType, err := p.extractParamType(pkgName, param)
			//if err != nil {
			//	return nil, nil, err
			//}
			paramSchemaRef, err := p.generateSchemaRef(param.Type)
			if err != nil {
				return nil, nil, err
			}
			propName := param.Names[0].Name
			propName = strings.ToLower(propName[:1]) + propName[1:]
			requestBody.Content["application/json"] = &openapi3.MediaType{
				Schema: paramSchemaRef,
			}
		}
		requestBodyRef.Value = requestBody
	}

	ep, ok := p.endpoints[operationID]
	if !ok {
		p.endpoints[operationID] = Endpoint{}
	}

	ep.RequestBodyTypes = *requestBodyRef
	//ep.ResponseBodyTypes = *responses

	p.endpoints[operationID] = ep
	return requestBodyRef, responses, err
}

func (p *parserClient) extractParamType(pkgName string, param *ast.Field) (string, error) {
	if param.Type == nil {
		return "", nil
	}
	switch t := param.Type.(type) {
	case *ast.SelectorExpr:
		identPkgName, ok := t.X.(*ast.Ident)
		if !ok {
			return "", fmt.Errorf("failed to extract package name from selector expression")
		}
		pkgPath, err := p.getPackageImportPath(identPkgName.Name)
		if err != nil {
			p.log("import Path is null, using as is...")
			//return "", fmt.Errorf("failed to get package import Path: %w", err)
		}
		typeName := t.Sel.Name
		return pkgPath + "." + typeName, nil
	case *ast.StarExpr:
		return p.extractType(pkgName, t.X)
	case *ast.ArrayType:
		typeStr, err := p.extractType(pkgName, t.Elt)
		if err != nil {
			return "", fmt.Errorf("failed to extract array element type: %w", err)
		}
		return "[]" + typeStr, nil
	default:
		return p.extractType(pkgName, t)
	}
}
func (p *parserClient) getPackageImportPath(pkgName string) (string, error) {
	pkg, ok := p.packages[pkgName]
	if !ok {
		p.log("package %s not found, using as is...", pkgName)
		return pkgName, nil
	}

	if len(pkg.Files) == 0 {
		return "", fmt.Errorf("no files found for package %s", pkgName)
	}

	var file *ast.File
	for _, f := range pkg.Files {
		if f.Name.Name == "doc" {
			file = f
			break
		}
	}

	if file == nil {
		// if no file found with package name as filename, then check if only one file exists in package
		if len(pkg.Files) == 1 {
			for _, f := range pkg.Files {
				file = f
				break
			}
		}
		if file == nil {
			return "", fmt.Errorf("no file found for package %s", pkgName)
		}
	}

	for _, i := range file.Imports {
		importPath := strings.Trim(i.Path.Value, `"`)
		if i.Name != nil {
			if i.Name.Name == pkgName {
				return importPath, nil
			}
		} else {
			// no alias, use last part of import Path
			parts := strings.Split(importPath, "/")
			if parts[len(parts)-1] == pkgName {
				return importPath, nil
			}
		}
	}

	return "", fmt.Errorf("package %s not found in any imports", pkgName)
}

//func (p *parserClient) getPackageImportPath(identPkgName string) string {
//	// Search for the package among already processed ones.
//	fmt.Printf("Indent Package: %s", identPkgName)
//	for _, pkg := range p.packages {
//		for _, ident := range pkg.Imports {
//			fmt.Printf("Cached Package: %s", ident.Name)
//			if strings.Contains(ident.Name, identPkgName) {
//				return ident.Name
//			}
//		}
//	}
//
//	return ""
//}

// extractType extracts the type of an expression and returns it as a string.
func (p *parserClient) extractType(pkgName string, typ ast.Expr) (string, error) {
	switch t := typ.(type) {
	case *ast.Ident:
		// If the type is an identifier, it's a basic type.
		return t.Name, nil
	case *ast.ArrayType:
		// If the type is an array, extract its element type.
		elemType, err := p.extractType(pkgName, t.Elt)
		if err != nil {
			return "", err
		}
		return "[]" + elemType, nil
	case *ast.StarExpr:
		// If the type is a pointer, extract the pointed-to type.
		pointedToType, err := p.extractType(pkgName, t.X)
		if err != nil {
			return "", err
		}
		return "*" + pointedToType, nil
	case *ast.MapType:
		// If the type is a map, extract its key and value types.
		keyType, err := p.extractType(pkgName, t.Key)
		if err != nil {
			return "", err
		}
		valueType, err := p.extractType(pkgName, t.Value)
		if err != nil {
			return "", err
		}
		return "map[" + keyType + "]" + valueType, nil
	case *ast.StructType:
		// If the type is a struct, extract its fields and types.
		fields := []string{}
		for _, field := range t.Fields.List {
			fieldType, err := p.extractType(pkgName, field.Type)
			if err != nil {
				return "", err
			}
			// If the field has a name, use it, otherwise use the type.
			if len(field.Names) > 0 {
				fields = append(fields, field.Names[0].Name+":"+fieldType)
			} else {
				fields = append(fields, fieldType)
			}
		}
		return "struct {" + strings.Join(fields, ";") + "}", nil
	case *ast.InterfaceType:
		// If the type is an interface, return "interface{}".
		return "interface{}", nil
	case *ast.SelectorExpr:
		// If the type is a selector expression, resolve the package and extract the type. t.X.(*ast.Ident).Name
		typedName, err := p.getPackage(pkgName)
		if err != nil {
			return "", err
		}
		return typedName.Name, nil
	default:
		// Otherwise, return an error.
		return "", fmt.Errorf("unsupported type: %T", typ)
	}
}

func (p *parserClient) extractNamedType(pkgName string, typeName string) (*types.Named, error) {
	pkg, err := p.getPackage(pkgName)
	if err != nil {
		return nil, err
	}
	obj := pkg.Scope.Lookup(typeName)
	if obj == nil {
		return nil, fmt.Errorf("Type %s not found in package %s", typeName, pkgName)
	}
	named, ok := obj.Type.(*types.Named)
	if !ok {
		return nil, fmt.Errorf("Type %s is not a named type in package %s", typeName, pkgName)
	}
	return named, nil
}

func (p *parserClient) extractSelectorType(pkgName string, t *ast.SelectorExpr) (string, error) {
	// extract the package name or nested expression
	expr, err := p.extractExprType(pkgName, t.X)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s", expr, t.Sel.Name), nil
}

// extractExprType extracts the type string from the expression.
func (p *parserClient) extractExprType(pkgName string, expr ast.Expr) (string, error) {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name, nil
	case *ast.SelectorExpr:
		return p.extractSelectorType(pkgName, t)
	case *ast.ArrayType:
		elementType, err := p.extractExprType(pkgName, t.Elt)
		if err != nil {
			return "", err
		}
		return "[]" + elementType, nil
	case *ast.StarExpr:
		pointerType, err := p.extractExprType(pkgName, t.X)
		if err != nil {
			return "", err
		}
		return "*" + pointerType, nil
	case *ast.MapType:
		keyType, err := p.extractExprType(pkgName, t.Key)
		if err != nil {
			return "", err
		}
		valueType, err := p.extractExprType(pkgName, t.Value)
		if err != nil {
			return "", err
		}
		return "map[" + keyType + "]" + valueType, nil
	case *ast.InterfaceType:
		return "interface{}", nil
	default:
		return "", fmt.Errorf("unable to extract type from expr: %#v", expr)
	}
}
func (p *parserClient) generateSchemaRef(respType ast.Expr) (*openapi3.SchemaRef, error) {
	p.log("generateSchemaRef")
	schema, err := p.generateSchema(respType)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

func (p *parserClient) generateSchema(typ ast.Expr) (*openapi3.SchemaRef, error) {
	p.log("generateSchemaRef")
	switch t := typ.(type) {
	case *ast.Ident:
		return p.basicTypeToSchemaRef(t.Name)
	case *ast.StarExpr:
		schemaRef, err := p.generateSchemaRef(t.X)
		if err != nil {
			return nil, err
		}
		//schemaRef.Value.Nullable = true
		return schemaRef, nil
	case *ast.ArrayType:
		itemsRef, err := p.generateSchema(t.Elt)
		if err != nil {
			return nil, err
		}
		return &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:  "array",
				Items: itemsRef,
			},
		}, nil
	case *ast.MapType:
		additionalPropsRef, err := p.generateSchemaRef(t.Value)
		if err != nil {
			return nil, err
		}
		return &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: "object",
				AdditionalProperties: openapi3.AdditionalProperties{
					Schema: additionalPropsRef,
				},
			},
		}, nil
	case *ast.SelectorExpr:
		// Handle imported package types
		pkgName := t.X.(*ast.Ident).Name
		typeName := t.Sel.Name
		importPath, err := p.getPackageImportPath(pkgName)
		if err != nil {
			return nil, err
		}
		ref := fmt.Sprintf("%s#/components/schemas/%s", importPath, typeName)
		return &openapi3.SchemaRef{
			Ref: ref,
		}, nil
	case *ast.StructType:
		schema, err := p.generateSchema(t)
		if err != nil {
			return nil, err
		}
		return schema, nil
	default:
		return nil, fmt.Errorf("unsupported type %T", t)
	}
}

func (p *parserClient) basicTypeToSchemaRef(typeName string) (*openapi3.SchemaRef, error) {
	switch typeName {
	case "string":
		return openapi3.NewStringSchema().NewRef(), nil
	case "int", "int32":
		return openapi3.NewIntegerSchema().NewRef(), nil
	case "int64":
		return openapi3.NewIntegerSchema().WithFormat("int64").NewRef(), nil
	case "float64":
		return openapi3.NewFloat64Schema().WithFormat("float").NewRef(), nil
	case "bool":
		return openapi3.NewBoolSchema().NewRef(), nil
	case "time.Time":
		return openapi3.NewStringSchema().WithFormat("date-time").NewRef(), nil
	default:
		return nil, fmt.Errorf("unsupported basic type %s", typeName)
	}
}

func (p *parserClient) generatePathItem(summary, description string, method string, operation *openapi3.Operation) *openapi3.PathItem {
	pathItem := openapi3.PathItem{}

	switch method {
	case "GET":
		pathItem.Get = operation
	case "POST":
		pathItem.Post = operation
	case "PUT":
		pathItem.Put = operation
	case "DELETE":
		pathItem.Delete = operation
	case "PATCH":
		pathItem.Patch = operation
	case "HEAD":
		pathItem.Head = operation
	case "OPTIONS":
		pathItem.Options = operation
	}

	if summary != "" {
		pathItem.Summary = summary
	}

	if description != "" {
		pathItem.Description = description
	}

	return &pathItem
}

func (p *parserClient) resolveBuiltinType(name string) (string, error) {
	switch name {
	case "bool":
		return "boolean", nil
	case "string":
		return "string", nil
	case "int", "int8", "int16", "int32", "uint", "uint8", "uint16", "uint32":
		return "integer", nil
	case "int64", "uint64":
		return "integer", fmt.Errorf("64-bit integer type %q not supported", name)
	case "float32", "float64":
		return "number", nil
	case "byte", "rune", "uintptr", "complex64", "complex128", "error":
		return "", fmt.Errorf("builtin type %q not supported", name)
	default:
		return "", fmt.Errorf("unknown builtin type %q", name)
	}
}

func (p *parserClient) extractEndpointParameters(pkgName string, fn *ast.FuncDecl) (*openapi3.SchemaRef, *openapi3.SchemaRef, error) {
	inputParams := fn.Type.Params.List
	var paramsSchema *openapi3.SchemaRef
	var requestSchema *openapi3.SchemaRef

	// extract request body
	if len(inputParams) == 1 {
		param := inputParams[0]
		if isHTTPRequestBody(param) {
			name := param.Names[0].Name
			typ := param.Type
			schema, err := p.extractType(pkgName, typ)
			if err != nil {
				return nil, nil, fmt.Errorf("error extracting parameter '%s' type: %w", name, err)
			}
			requestSchema = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: schema,
				},
			}
		}
	}

	// extract Parameters
	if len(inputParams) > 1 {
		params := make([]*openapi3.Parameter, 0, len(inputParams))
		for _, param := range inputParams {
			if !isHTTPRequestBody(param) {
				name := param.Names[0].Name
				typ := param.Type
				schema, err := p.extractType(pkgName, typ)
				if err != nil {
					return nil, nil, fmt.Errorf("error extracting parameter '%s' type: %w", name, err)
				}
				params = append(params, &openapi3.Parameter{
					Name:     name,
					In:       "query",
					Required: !isOptional(param),
					Schema: &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: schema,
						},
					},
				})
			}
		}
		paramsSchema = &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:       "object",
				Properties: make(map[string]*openapi3.SchemaRef),
			},
		}

		for _, param := range params {
			paramsSchema.Value.Properties[param.Name] = param.Schema
		}
	}

	return paramsSchema, requestSchema, nil
}

func (p *parserClient) extractParamSchema(param *openapi3.SchemaRef) (*openapi3.SchemaRef, error) {
	// If the schema is already resolved or is not of type object, return it as is
	if param.Value != nil && param.Value.Type != "object" {
		return param, nil
	}

	if param.Ref != "" {
		// If the schema is a reference to another schema, resolve the reference
		refSchema, err := p.resolveRef(param.Ref)
		if err != nil {
			return nil, err
		}
		param = refSchema
	} else if len(param.Value.Properties) > 0 {
		// If the schema has properties, use it as is
		return param, nil
	} else {
		// If the schema has no properties, try to extract its type and generate a new schema with the type
		schemaType, err := p.extractType("", &ast.Ident{Name: param.Value.Type})
		if err != nil {
			return nil, err
		}
		//param, err = p.generateSchemaRef(param.Value.Type)
		p.log("// TODO: failed to get any schema for  %s", schemaType)
	}

	return param, nil
}

func (p *parserClient) resolveRef(ref string) (*openapi3.SchemaRef, error) {
	components := p.doc.Components
	if components == nil {
		return nil, fmt.Errorf("OpenAPI document does not have a 'components' section")
	}

	// Split the ref into its components
	parts := strings.Split(ref, "/")
	if len(parts) < 2 || parts[0] != "#/components/schemas" {
		return nil, fmt.Errorf("invalid schema reference: %s", ref)
	}

	schemaName := parts[1]

	schema, ok := components.Schemas[schemaName]
	if !ok {
		return nil, fmt.Errorf("schema %s not found in document components", schemaName)
	}

	// Return a new reference that points directly to the schema object
	return schema, nil
}

func isHTTPRequestBody(param *ast.Field) bool {
	_, ok := param.Type.(*ast.StarExpr)
	if !ok {
		return false
	}

	selector, ok := param.Type.(*ast.StarExpr).X.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if selector.Sel.Name != "Request" {
		return false
	}

	return true
}

func isOptional(param *ast.Field) bool {
	if len(param.Names) == 0 {
		return false
	}

	_, isVariadic := param.Type.(*ast.Ellipsis)
	if isVariadic {
		return true
	}

	return param.Tag != nil && strings.Contains(param.Tag.Value, "optional")
}
