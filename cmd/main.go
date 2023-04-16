package main

import (
	"flag"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/imdario/mergo"
	"github.com/vasusheoran/go-openapi/scan"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

func main() {
	var dir, output, level, f, meta string
	flag.StringVar(&dir, "dir", ".", "the directory containing the Go files to parse")
	flag.StringVar(&level, "level", "meta", "sets the logging level. default is meta")
	flag.StringVar(&output, "output", "./openapi.yaml", "the file path where the OpenAPI specification file will be written, default is 'openapi.yaml'")
	flag.StringVar(&f, "file", "", "the file path to override generated file")
	flag.StringVar(&meta, "meta", "", "the file path that OpenAPI meta relative to the dir")
	flag.Parse()

	if dir == "" {
		dir = "openapi.yaml"
	}

	spec, err := generateSpec(dir, level, meta)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	err = writeSpec(spec, output, f)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	fmt.Println("OpenAPI specification generated successfully!")
}

func generateSpec(dir, level, meta string) (*openapi3.T, error) {
	return scan.NewParser().WithLogLevel(level).WithMetaPath(meta).GetSpec(dir)
}

func writeSpec(spec *openapi3.T, output string, f string) error {
	if len(f) > 0 {
		// TODO
		data, err := ioutil.ReadFile(f)
		if err != nil {
			panic(err)
		}
		spec2, err := openapi3.NewLoader().LoadFromData(data)
		if err != nil {
			panic(err)
		}

		err = mergo.Merge(spec, *spec2)
		if err != nil {
			fmt.Println(err)
		}
	}

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
