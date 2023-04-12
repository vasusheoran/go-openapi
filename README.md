# go-openapi
go-openapi is a Go program that generates an OpenAPI 3.0 specification for a given Go application. It utilizes the go-swagger package to extract information about the application's routes and endpoints, and converts this information into a valid OpenAPI 3.0 specification.
## Installation
To install go-openapi, use the following command:
```shell
go build cmd/main.go
```
## Usage
To generate an OpenAPI specification for your Go application, run the following command:
```shell
go-openapi --dir <path/to/dir> --output </path/to/output.yaml>
```
The `output` argument specifies the file path where the generated specification file will be saved. 
The `dir` argument specifies the project dir that will be scanned.
## Contributing
If you would like to contribute to go-openapi, please feel free to submit a pull request with your changes.
## License
go-openapi is released under the MIT License. See LICENSE for more information.
