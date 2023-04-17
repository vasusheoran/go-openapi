# go-openapi
go-openapi is a Go program that generates an OpenAPI 3.0 specification for a given Go application. It utilizes the go parser to extract information about the application's routes and endpoints, and converts this information into a valid OpenAPI 3.0 specification.
## Installation
To install go-openapi, use the following command:
```shell
go build cmd/main.go -o go-openapi
```
## Usage
To generate an OpenAPI specification for your Go application, run the following command:
```shell
go-openapi [Options]
```
| Options  | Description                                                                                                                                  |
|----------|----------------------------------------------------------------------------------------------------------------------------------------------|
| `dir`    | A comma-separated list of directories to be scanned.                                                                                         |
| `output` | The file path for the generated spec.                                                                                                        |
| `values` | A comma-separated list of OpenAPI 3.1 compliant specifications to be merged into the generated spe                                           |
| `meta`   | An optional field to specify the file path for metadata from the scanned directories in case multiple files contain openapi:meta annotation. |
| `level`  | The logging level. The default value is set to Info.                                                                                         |

### openapi.yaml generation
The toolkit has a command that will let you generate a OAS 3.1 spec document from your code. The command integrates with go doc comments, and 
makes use of structs when it needs to know of types.

Based on the work from https://github.com/go-swagger/go-swagger.

It uses a similar approach but with expanded annotations and it produces a Open API 3.0 spec.

The goal of the syntax is to make it look as a natural part of the documentation for the application code.

The generator is passed a list of directories in order of dependencies and it uses that to discover all the code in use. To do this it makes use of go's parser package.

Once the parser has encountered a comment that matches one of its known tags, the parser extracts the relevant info from the comment. Currently parser does not support multiline comments.
### openapi:meta
The openapi:meta annotation flags a file as source for metadata about the API. This is typically a main.go file with your package documentation.

Server, tag can be specified here. The description property uses the rest of the comment block as description for the api when not explicitly provided.

```shell
openapi:meta
```
| Field                          | Description                                               |
|--------------------------------|-----------------------------------------------------------|
| `info title [value]`           | The title for the REST API generated spec.                |
| `info description start`       | The start annotation for the description of the REST API. |
| `info description end`         | The end annotation for the description of the REST API.   |
| `info version [value]`         | The version of the generated spec.                        |
| `servers [Value] [Value] ...`  | The hosts from where the spec is served.                  |
| `tag <Name> --- <Description>` | A grouping operation under the same tag.                  |
| `contact <URL> <Name>`         | Contact information about the generated spec.             |

```go
// openapi:meta info title Application protection REST API
// openapi:meta info description start
// Application protection manages data protection of applications.
// openapi:meta info description end
// openapi:meta info version v1
// openapi:meta server https://localhost:8080 https://localhost:8081
// openapi:meta tag Host Management --- Everything about your pets
// openapi:meta contact https://mysupport.netapp.com NetApp Support

package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, world!")
}

```
### openapi:operation
A openapi:operation annotation links a path to a method. This operation gets a unique id, which is used in various places. One such usage is in method names for client generation for example.

Because there are many routers available, this tool does not try to parse the paths you provided to your routing library of choice. So you have to specify your path pattern yourself in valid Open API 3.1 (YAML) syntax.
```shell
openapi:operation [Method] [Path] [OperationID]
```

You can find all the properties at https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md


| Field                                        | Description                                                                                                                                     |
|----------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------|
| `summary [value]`                            | A brief summary of the operation.                                                                                                               |
| `description [value]`                        | A detailed description of the operation.                                                                                                        |
| `tag [value] [value]`                        | A tag or set of tags that can be used to group related operations together.                                                                     |
| `produce [MediaType] [MediaType`             | The expected response media types for the operation. Each media type should be separated by a space.                                            |
| `consumes [MediaType] [MediaType]`           | The expected request media types for the operation. Each media type should be separated by a space.                                             |
| `param [Name] [In] [Object] [Required]`      | Describes a single parameter for the operation, including its name, location (e.g., query, path), data type, and whether it is required.        |
| `response [Code] [Object] --- [Description]` | Describes a possible response for the operation, including the HTTP status code, the response object, and a brief description of the response.  | 

```go

// PetsInterface This is a sample interface comment
// Interface are used to create tags. They must have `name` annotation associated with them.
type PetsInterface interface {
    // GetPets Fetches pets
    // openapi:operation GET /pets/{petID} getPet
    // openapi:summary Fetches pet
    // openapi:description Fetches all pet
    // openapi:tag pets
    // openapi:consumes application/json
    // openapi:produces application/json
    // openapi:param name param string true --- Name to filter pets
    // openapi:param petID path string true --- PetID to fetch Pet
    // openapi:param x-agent-id header string true --- Agent ID for the request
    // openapi:response 200 GetAllPets --- Response for GetPetByID API
    // openapi:response 400 ErrorResponse --- Error
    GetPet(id, name string) (GetPets, error)
}
```


### openapi:schema
```shell
openapi:schema [Name]
```
A openapi:schema annotation optionally gets a model name as extra data on the line. When this appears anywhere in a comment for a struct, then that struct becomes a schema in the definitions object of OpenAPI.
The struct gets analyzed and all the collected models are added to the tree. 

Definitions will appear in the generated spec if tagged with schema, whether they are actually used somewhere or not in the application. 

The fields are tracked separately so that they can be renamed later on using `openapi:name` tag with the field.
#### Fields

| Field                       | Description                                                                                                                                                                                   |
|-----------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `format [value]`            | Format for the field (e.g. date-time, uri, email, etc.)                                                                                                                                       |
| `nullable`                  | Boolean to represent if the field is nullable or not                                                                                                                                          |
| `example [value]`           | Example value for the field                                                                                                                                                                   |
| `required`                  | Boolean to represent if the field is a required field or not                                                                                                                                  |
| `oneOf [Value] [Value] ...` | Annotation for fields that should have one of the values mentioned in the OpenAPI Specification (OAS) 3.1, regardless of the field's type in the struct. Field type in the struct is ignored. |
| `name [Name]`               | Optional annotation for the name of the generated field. Use this in case the field name is different than the generated schema name.                                                         |
| `enum [Value] [Value] ...`  | Annotation to include enums for the field.                                                                                                                                                    |

```go

// Category ...
// openapi:schema category
type Category struct {
    // openapi:description Pet ID
    // openapi:example 1
    // openapi:default 1
    ID int `json:"id"`
    // openapi:description Category name for the pets
    // openapi:example dog
    // openapi:nullable
    // openapi:default cat
    // openapi:enum cat dog
    Name string `json:"name"`
    // openapi:name category
    // openapi:oneOf Cat Dog
    Category json.RawMessage `json:"category"`
}
```
## Contributing
If you would like to contribute to go-openapi, please feel free to submit a pull request with your changes.
## License
go-openapi is released under the MIT License. See LICENSE for more information.
