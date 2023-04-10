package openapi

import (
	"fmt"
	"testing"
)

func TestParserClient_ParseComments(t *testing.T) {
	svc := NewParser()
	pkg, err := svc.GenerateOpenAPI("testdata")
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(pkg)
}
