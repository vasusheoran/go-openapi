# go-openapi
go-openapi is a Go program that generates an OpenAPI 3.0 specification for a given Go application. It utilizes the go-swagger package to extract information about the application's routes and endpoints, and converts this information into a valid OpenAPI 3.0 specification.
## Installation
To install go-openapi, use the following command:
```shell
go get github.com/{YOUR_GITHUB_USERNAME}/go-openapi
```
## Usage
To generate an OpenAPI specification for your Go application, run the following command:
```shell
go-openapi <output_directory> <output_format>
```
The `output_directory` argument specifies the directory where the generated specification file(s) will be saved. The `output_format` argument specifies the desired format for the output file(s), which can be either `json` or `yaml`.
## Contributing
If you would like to contribute to go-openapi, please feel free to submit a pull request with your changes.
## License
go-openapi is released under the MIT License. See LICENSE for more information.
