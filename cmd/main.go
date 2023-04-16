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
	"strings"
)

type overrideSpecSlice []string

func (s *overrideSpecSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *overrideSpecSlice) Set(value string) error {
	valuesSlice := strings.Split(value, ",")

	for _, v := range valuesSlice {
		*s = append(*s, v)

	}
	return nil
}

var logger = scan.NewLogger(scan.LogLevelInfo)
var dir, output, level, meta string
var values overrideSpecSlice

func main() {
	flag.StringVar(&dir, "dir", ".", "the directory containing the Go files to parse")
	flag.StringVar(&level, "level", "", "sets the logging level. default is `info`")
	flag.StringVar(&output, "output", "./openapi.yaml", "the file path where the OpenAPI specification file will be written, default is 'openapi.yaml'")
	flag.Var(&values, "values", "comma separated list of override spec files")
	flag.StringVar(&meta, "meta", "", "the file path that OpenAPI meta relative to the dir")
	flag.Parse()

	if dir == "" {
		dir = "openapi.yaml"
	}

	if len(level) != 0 {
		switch strings.ToLower(level) {
		case "debug":
			logger.SetLogLevel(scan.LogLevelDebug)
		case "info":
			logger.SetLogLevel(scan.LogLevelInfo)
		case "warn":
			logger.SetLogLevel(scan.LogLevelWarn)
		case "error":
			logger.SetLogLevel(scan.LogLevelError)
		case "fatal":
			logger.SetLogLevel(scan.LogLevelFatal)
		default:
			logger.Warn("unsupported log level %s", level)
		}
	}

	spec, err := generateSpec()
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	spec, err = mergeSpec(spec)
	if err != nil {
		logger.Fatal("failed to merge spec")
	}

	err = writeSpec(spec)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	logger.Info("OpenAPI 3.0.1 specification `%s` generated successfully! Your API documentation is now up-to-date. Thank you for using our program!", output)
}

func generateSpec() (*openapi3.T, error) {
	return scan.NewParser(logger).WithMetaPath(meta).GetSpec(dir)
}

func mergeSpec(spec *openapi3.T) (*openapi3.T, error) {
	if len(values) == 0 {
		return nil, nil
	}

	logger.Info("Merging OpenAPI specification files %s with generated specs into %s...", values, output)
	for _, value := range values {
		data, err := ioutil.ReadFile(value)
		if err != nil {
			return nil, err
		}
		spec2, err := openapi3.NewLoader().LoadFromData(data)
		if err != nil {
			return nil, err
		}

		err = mergo.Merge(spec, *spec2)
		if err != nil {
			return nil, err
		}
	}

	return spec, nil
}

func writeSpec(spec *openapi3.T) error {

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
