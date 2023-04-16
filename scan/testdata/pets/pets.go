// openapi:meta info title Swagger Petstore - OpenAPI 3.1
// openapi:meta info description start
//
//	This is a sample Pet Store server based on the OpenAPI 3.1 specification.  You can find out more about \nSwagger at [https://swagger.io](https://swagger.io). In the third iteration of the pet store, we've switched to the design first approach! \nYou can now help us improve the API whether it's by making changes to the definition itself or to the code. \nThat way, with time, we can improve the API in general, and expose some of the new features in OAS3.
//
//	Some useful links:
//	- [The Pet Store repository](https://github.com/swagger-api/swagger-petstore)
//	- [The source API definition for the Pet Store](https://github.com/swagger-api/swagger-petstore/blob/master/src/main/resources/openapi.yaml)
//
// openapi:meta info description end
// openapi:meta info version 1.0.0
// openapi:meta info oas 3.1.0
// openapi:meta server https://localhost:8080 https://localhost:8081
// openapi:meta tag Pets Management  --- Everything about your pets
// openapi:meta contact https://mysupport.github.com GitHub Support

package main

import "encoding/json"

// CreatePetResponse ...
// openapi:schema
// openapi:xml create-pet
// TODO: handle oneOf
// TODO: add support for go validator for enums and regex ?
// TODO: handle root user for xml
type CreatePetResponse struct {
	// This is a sample field comment
	// openapi:description Returns ID for the per
	// openapi:format text
	// openapi:default "12-sdf-1-321"
	// openapi:example "12-sdf-1-321"
	ID string `json:"id"`
}

// GetPetByIDResponse ...
type GetPetByIDResponse struct {
	// openapi:description Name of the pet
	// openapi:example "rambo"
	// openapi:nullable
	// openapi:format text
	// openapi:default "tommy"
	Name string `json:"name"`
	// This is a sample field comment
	// openapi:description Type of pet
	// openapi:nullable
	Category json.RawMessage `json:"category"`
}

// GetPets ...
type GetPets struct {
	// openapi:description Returns list of pets
	// openapi:format array
	Pets []GetPetByIDResponse `json:"pets"`
}

// Category ...
// openapi:schema
// openapi:xml category
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
}

// Dog ...
// openapi:schema
type Dog struct {
	// openapi:description Name of the pet
	// openapi:example rambo
	// openapi:default tommy
	Name string `json:"name"`
}

// Cat ...
// openapi:schema
type Cat struct {
	// openapi:description Name of the pet
	// openapi:example rambo
	// openapi:default tommy
	Name string `json:"name"`
	// openapi:description Owner of the pet
	// openapi:example person1
	// openapi:default person1
	Owner string `json:"owner"`
}

// CreatePetRequest ...
// openapi:schema
// openapi:xml pet-request
type CreatePetRequest struct {
	// openapi:description Pet ID
	// openapi:example 1
	// openapi:default 1
	Id int `json:"id"`
	// Note that fields do not require openapi annotations to be parsed, that is must for strcuts, interfaces and methods.
	// All the nested objects will be parsed recursively
	// This is a sample field comment
	// openapi:description Type of the pet
	// openapi:nullable
	// Note that field type is ignored for the oneOf schemas and a references are injected as per openapi:oneOf
	// openapi:oneOf Cat Dog
	Type Category `json:"type"`
}

// PetsInterface This is a sample interface comment
// Interface are used to create tags. They must have `name` annotation associated with them.
type PetsInterface interface {
	// CreatePet Add a new pet to the store
	// openapi:operation POST /pets createPet
	// openapi:summary Adds a new pet to the store
	// openapi:description Adds a new pet to the store
	// openapi:tag pets
	// openapi:consumes application/json application/xml
	// openapi:produces application/json application/xml
	// openapi:param name query string false --- Name of pet that needs to be updated
	// openapi:param petId path string true --- ID of pet that needs to be updated
	// openapi:param x-agent-id header string true --- Agent ID for the request
	// openapi:body CreatePetRequest --- Request body to create Pets
	// openapi:response 200 CreatePetResponse --- Response for CreatePet API
	// openapi:response 200 CreatePetResponse --- OK
	// openapi:response 400 ErrorResponse --- Error
	CreatePet(name string) (*CreatePetResponse, error)
	GetPetByID(petId string) (GetPetByIDResponse, error)
	GetPets() error
}

type StoreInterface interface {
	CreateStore() error
	GetStore(id string) error
}

func main() {}
