package main

import (
	"os"
	"testing"
)

func Test_main_1(t *testing.T) {
	os.Args = []string{"cmd", "--inputDir", "../scan/testdata/v1.go", "--outputDir", "."}
	main()
}

func Test_main_2(t *testing.T) {
	os.Args = []string{"cmd", "--inputDir", "../scan/testdata/v2.go", "--output", "./v2.yaml"}
	main()
}

func Test_main_3(t *testing.T) {
	os.Args = []string{"cmd", "--inputDir", "../scan/testdata/v2.go", "--output", "./v3.json", "--outputType", "json"}
	main()
}
