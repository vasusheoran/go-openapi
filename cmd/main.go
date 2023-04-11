package main

import (
	"flag"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/vasusheoran/go-openapi/scan"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"path/filepath"
)

func main() {
	var inputDir, output, outputType string
	flag.StringVar(&inputDir, "inputDir", "", "the directory containing the Go files to parse")
	flag.StringVar(&output, "output", ".", "the directory where the OpenAPI specification file will be written, default is '.'")
	flag.StringVar(&outputType, "outputType", "yaml", "the output file format (json or yaml), default is 'yaml'")
	flag.Parse()

	if inputDir == "" {
		log.Fatal("error: input directory is required")
	}

	if output == "" {
		output = filepath.Join(inputDir, "docs")
	}

	if outputType != "json" && outputType != "yaml" {
		log.Fatalf("error: unsupported output type %s", outputType)
	}

	spec, err := generateSpec(inputDir)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	err = writeSpec(spec, output, outputType)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	fmt.Println("OpenAPI specification generated successfully!")
}

func generateSpec(inputDir string) (*openapi3.T, error) {
	parser := scan.NewParser()
	err := parser.ParseFile(inputDir)
	if err != nil {
		return nil, err
	}

	return parser.GetSpec(), nil
}

func writeSpec(spec *openapi3.T, output, outputType string) error {
	var b []byte
	var err error

	//err = os.MkdirAll(output, os.ModePerm)
	//if err != nil {
	//	return err
	//}

	switch outputType {
	case "json":
		b, err = spec.MarshalJSON()
	case "yaml":
		b, err = yaml.Marshal(spec)
	default:
		return fmt.Errorf("unsupported output type %q", outputType)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal spec to %s: %w", outputType, err)
	}

	//err = os.MkdirAll(output, 0755)
	//if err != nil {
	//	return fmt.Errorf("failed to create output directory: %w", err)
	//}

	//filePath := filepath.Join(output, outputType)
	err = ioutil.WriteFile(output, b, 0644)
	if err != nil {
		return fmt.Errorf("failed to write spec file: %w", err)
	}

	return nil
}
