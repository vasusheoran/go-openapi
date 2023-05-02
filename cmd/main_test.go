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
	os.Args = []string{"cmd", "--dir", "../scan/testdata/pets", "--output", "openapi.yaml", "--meta", "pets.go", "--level", "debug"}
	//os.Args = []string{"cmd", "--dir", "../scan/testdata/pets1", "--output", "openapi.yaml", "--level", "debug"}
	main()
}

func Test_Host_Management(t *testing.T) {
	os.Args = []string{"cmd", "--dir", "/Users/vasusheoran/git/applicationsmicroservices/common-go,/Users/vasusheoran/git/applicationsmicroservices/host-manager", "--output", "openapi-hm.yaml", "--meta", "cmd/main.go"}
	main()
}
