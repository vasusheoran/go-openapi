// openapi:info title Swagger Petstore - OpenAPI 3.1
// openapi:info description start
//
//	This is a sample Pet Store Server based on the OpenAPI 3.1 specification.  You can find out more about \nSwagger at [https://swagger.io](https://swagger.io). In the third iteration of the pet store, we've switched to the design first approach! \nYou can now help us improve the API whether it's by making changes to the definition itself or to the code. \nThat way, with time, we can improve the API in general, and expose some of the new features in OAS3.
//
//	Some useful links:
//	- [The Pet Store repository](https://github.com/swagger-api/swagger-petstore)
//	- [The source API definition for the Pet Store](https://github.com/swagger-api/swagger-petstore/blob/master/src/main/resources/openapi.yaml)
//
// openapi:info description end
// openapi:info version 1.0.0
// openapi:info oas 3.1.0
// openapi:info server localhost:8080 localhost:8081
package main

// CreatePetResponse ...
type CreatePetResponse struct {
	// This is a sample field comment
	// openapi:description Returns ID for the per
	// openapi:format text
	// openapi:default "12-sdf-1-321"
	// openapi:example "12-sdf-1-321"
	ID string `json:"id"`
}

// GetPetByIDResponse ...
// TODO: handle oneOf
// TODO: add support for go validator for enums and regex ?
type GetPetByIDResponse struct {
	// openapi:description Name of the pet
	// openapi:example "rambo"
	// openapi:nullable
	// openapi:format text
	// openapi:default "tommy"
	Name string `json:"name"`
	// This is a sample field comment
	// openapi:description Type of pet
	// openapi:nullable true
	// openapi:format text
	Category Category `json:"category"`
}

type GetPets struct {
	// openapi:description Returns list of pets
	// openapi:format array
	Pets []GetPetByIDResponse `json:"pets"`
}

// Category ...
type Category struct {
	// openapi:description Pet ID
	// openapi:example "1"
	// openapi:format text
	// openapi:default "1"
	ID int `json:"id"`
	// openapi:description Category name for the pets
	// openapi:example "dog"
	// openapi:nullable true
	// openapi:format text
	// openapi:default "cat"
	// openapi:enum "cat,dog"
	Name string `json:"name"`
}

// CreatePetRequest ...
type CreatePetRequest struct {
	// openapi:description Pet ID
	// openapi:example "1"
	// openapi:format text
	// openapi:default "1"
	Id int `json:"id"`
	// openapi:description Name of the pet
	// openapi:example "rambo"
	// openapi:format text
	// openapi:default "tommy"
	Name string `json:"name"`
	// Note that fields do not require openapi annotations to be parsed, that is must for strcuts, interfaces and methods.
	// All the nested objects will be parsed recursively
	Category Category `json:"category"`
}

// PetsInterface This is a sample interface comment
// Interface are used to create tags. They must have `name` annotation associated with them.
// openapi:name pet
// openapi:description Everything about your Pets
// openapi:external-docs http://github.com/vasusheoran @author:vasusheoran
// TODO: Global Parameters
// TODO: OperationID required for struct
// TODO: If tags not present, then use the interface name by default
// TODO: Set op to method name by default
type PetsInterface interface {
	// CreatePet Add a new pet to the store
	// openapi:summary Adds a new pet to the store
	// openapi:description Adds a new pet to the store
	// openapi:tags pet
	// openapi:id CreatePet
	// openapi:path /pets
	// openapi:method POST
	// openapi:body CreatePetRequest
	// openapi:success 200 CreatePetResponse
	// openapi:failure 400 ErrorResponse
	CreatePet() (*CreatePetResponse, error)
	// GetPetByID This is a sample method 2 comment
	// openapi:summary Find pet by ID
	// openapi:description Returns a single pet
	// openapi:tags pet
	// openapi:id GetPetsOp
	// openapi:path /pets/{petId}
	// openapi:method GET
	// openapi:param name query string false "Name of pet that needs to be updated"
	// openapi:param petId path string true "ID of pet that needs to be updated"
	// openapi:param x-agent-id header string true "Agent ID for the request"
	// openapi:success 204
	// openapi:failure 400 ErrorResponse
	GetPetByID(petId string) (*GetPetByIDResponse, error)
	// GetPets Returns list of pets
	// openapi:summary Get all pets
	// openapi:description Returns list of all pets
	// openapi:tags pet
	// openapi:id GetPetByIDOp
	// openapi:path /pets
	// openapi:method GET
	// openapi:success 200 GetPets
	// openapi:failure 400 ErrorResponse
	GetPets() error
}

func main() {}
