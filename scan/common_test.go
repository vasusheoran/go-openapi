// comment before package
package scan

// comment after package
import (
	"fmt"
	"go/parser"
	"go/token"
	"testing"
)

func TestParser_extractRequestBodyFromComment(t *testing.T) {
	// Create a file set to keep track of position information for the parsed files
	fset := token.NewFileSet()

	// Parse the Go source file
	file, err := parser.ParseFile(fset, "common_test.go", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// Check if a package declaration was found
	if file.Package != token.NoPos {
		fmt.Println("Found a package declaration")
	} else {
		fmt.Println("No package declaration found")
	}
}
