package main

import (
	"os"
	"testing"
)

func Test_main_1(t *testing.T) {
	//os.Args = []string{"cmd", "--inputDir", "../scan/testdata/pets.go", "--output", "v2.yaml"}
	os.Args = []string{"cmd", "--dir", "../scan/testdata/pets", "--level", "warn"}
	main()
}

func Test_main_2(t *testing.T) {
	//os.Args = []string{"cmd", "--inputDir", "../scan/testdata/pets.go", "--output", "v2.yaml"}
	os.Args = []string{"cmd", "--dir", "../scan/testdata/info", "--level", "warn"}
	main()
}
