package scan

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestParser_ParseFile(t *testing.T) {
	svc := NewParser()
	err := svc.ParseFile("testdata/v1.go")
	assert.Nil(t, err)
}

func TestParser_Comments(t *testing.T) {
	//svc := NewParser()
	//err := svc.ParseFile("testdata/v2.go")
	//assert.Nil(t, err)
	printCommentsUniquely()
}

func printCommentsUniquely() {
	// Parse the source code
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "testdata/v2.go", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	var list []string
	comments := map[string][]string{}

	for _, decl := range node.Decls {
		switch declType := decl.(type) {
		case *ast.GenDecl:
			switch declType.Tok {
			case token.TYPE:
				// Handle type declarations
				for _, spec := range declType.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {

						intfDecl := decl.(*ast.GenDecl)
						for _, c := range intfDecl.Doc.List {
							list, ok = comments[ts.Name.Name]
							if !ok {
								list = []string{}
							}
							list = append(list, c.Text)
							comments[ts.Name.Name] = list
						}

						switch ts.Type.(type) {
						case *ast.InterfaceType:
							iface, ok := ts.Type.(*ast.InterfaceType)
							if !ok {
								break
							}

							for _, field := range iface.Methods.List {
								for _, c := range field.Doc.List {
									list, ok = comments[field.Names[0].Name]
									if !ok {
										list = []string{}
									}
									list = append(list, c.Text)
									comments[field.Names[0].Name] = list
								}
							}
						}
					}
				}
			case token.IMPORT:
				// Handle import declarations
			case token.CONST, token.VAR:
				// Handle const and var declarations
			}
		case *ast.FuncDecl:
			// Handle function declarations
		}
	}

	for key, list := range comments {
		for _, c := range list {
			println(key, ": ", c)
		}
	}
}

func printComments(f *ast.File, ts *ast.TypeSpec, fs *token.FileSet) []*ast.CommentGroup {
	iface, ok := ts.Type.(*ast.InterfaceType)
	if !ok {
		return nil
	}

	for _, field := range iface.Methods.List {
		for _, comment := range field.Doc.List {
			fmt.Println(comment.Text, " : ", field.Names[0].Name)
		}
	}
	//cg := extractComments(fs, iface)
	//for i, group := range cg {
	//	for _, comment := range group.List {
	//		//fmt.Println(comment.Text, " : ", i)
	//		fmt.Println(comment.Text, " : ", iface.Methods.List[i-1].)
	//	}
	//}
	return nil
}

func getCommentsForNode(commentMap []*ast.CommentGroup, node ast.Node) []*ast.Comment {
	var comments []*ast.Comment

	// Iterate over the comments in the comment map
	for _, group := range commentMap {
		for _, comment := range group.List {
			if comment.Pos() >= node.Pos() && comment.End() <= node.End() {
				comments = append(comments, comment)
			}
		}
	}

	return comments
}
