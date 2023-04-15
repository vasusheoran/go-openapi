package main

import (
	"os"
	"testing"
)

func Test_main_merge(t *testing.T) {
	//os.Args = []string{"cmd", "--inputDir", "../scan/testdata/pets2.go", "--output", "v2.yaml"}
	os.Args = []string{"cmd", "--dir", "../scan/testdata/pets", "--level", "warn", "--file", "../scan/testdata/pets/override.yaml", "--output", "openapi-merged.yaml"}
	main()
}

func Test_main_2(t *testing.T) {
	//os.Args = []string{"cmd", "--inputDir", "../scan/testdata/pets2.go", "--output", "v2.yaml"}
	os.Args = []string{"cmd", "--dir", "../scan/testdata/pets", "--level", "warn", "--output", "openapi.yaml"}
	main()
}
