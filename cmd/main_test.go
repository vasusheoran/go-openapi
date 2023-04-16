package main

import (
	"os"
	"testing"
)

func Test_merge(t *testing.T) {
	os.Args = []string{"cmd", "--dir", "../scan/testdata/pets", "--values", "../scan/testdata/pets/override.yaml", "--output", "openapi-merged.yaml"}
	main()
}

func Test_parser(t *testing.T) {
	os.Args = []string{"cmd", "--dir", "../scan/testdata/pets", "--output", "openapi.yaml", "--meta", "pets.go"}
	main()
}
