package main

import (
	"flag"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/vasusheoran/go-openapi/scan"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

func main() {
	var dir, output, level string
	flag.StringVar(&dir, "dir", ".", "the directory containing the Go files to parse")
	flag.StringVar(&level, "level", "info", "sets the logging level. default is info.")
	flag.StringVar(&output, "output", "./openapi.yaml", "the file path where the OpenAPI specification file will be written, default is 'openapi.yaml'")
	flag.Parse()

	if dir == "" {
		dir = "openapi.yaml"
	}

	spec, err := generateSpec(dir, level)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	err = writeSpec(spec, output)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	fmt.Println("OpenAPI specification generated successfully!")
}

func generateSpec(dir, level string) (*openapi3.T, error) {
	return scan.NewParser().WithLogLevel(level).GetSpec(dir)
}

func writeSpec(spec *openapi3.T, output string) error {
	//b, err := spec.MarshalJSON()
	//if err != nil {
	//	return fmt.Errorf("failed to marshal spec: %s", err.Error())
	//}
	b, err := yaml.Marshal(spec)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(output, b, 0644)
	if err != nil {
		return fmt.Errorf("failed to write spec file: %s", err.Error())
	}

	return nil
}
