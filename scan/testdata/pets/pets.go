package main

// CreatePetResponse ...
type CreatePetResponse struct {
	// This is a sample field comment
	// openapi:description Contains list of pet name
	// openapi:nullable true
	// openapi:format array
	// TODO: handle openapi:examples and array types for fields
	// openapi:default [a,b,c]
	// openapi:example "[a,b,c]"
	Name []string `json:"name"`
}

// GetPetByIDResponse ...
type GetPetByIDResponse struct {
	// openapi:description Name of the pet
	// openapi:example "rambo"
	// openapi:nullable true
	// openapi:format text
	// openapi:default "tommy"
	Name string `json:"name"`
	// This is a sample field comment
	// openapi:description Type of pet
	// openapi:nullable true
	// openapi:format text
	// TODO: handle openapi:oneOf Dog Cat
	Category Category `json:"category"`
}

// Category ...
type Category struct {
	// openapi:description Pet ID
	// openapi:example "1"
	// openapi:nullable false
	// openapi:format text
	// openapi:default "1"
	ID int `json:"id"`
	// openapi:description Category name for the pets
	// openapi:example "dog"
	// openapi:nullable true
	// openapi:format text
	// openapi:default "cat"
	// openapi:enum "cat,dog"
	// TODO: add support for go validator for enums and regex ?
	Name string `json:"name"`
}

// CreatePetRequest ...
type CreatePetRequest struct {
	// openapi:description Pet ID
	// openapi:example "1"
	// openapi:nullable false
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
// Headers
type PetsInterface interface {
	// CreatePet Add a new pet to the store
	// openapi:summary Adds a new pet to the store
	// openapi:description Adds a new pet to the store
	// openapi:tags pet
	// openapi:id CreatePet
	// openapi:path /pet
	// openapi:method GET
	// openapi:body SampleRequest1
	// openapi:success 200 CreatePetResponse
	// openapi:failure 400 ErrorResponse
	CreatePet() (*CreatePetResponse, error)
	// GetPetByID This is a sample method 2 comment
	// openapi:summary Find pet by ID
	// openapi:description Returns a single pet
	// TODO: If tags not present, then use the interface name by default
	// openapi:tags pet1
	// TODO: Set op to method name by default
	// openapi:id GetPetByIDOp
	// openapi:path /pet/{petId}
	// openapi:method GET
	// openapi:param name query string false "Name of pet that needs to be updated"
	// openapi:param petId path string true "ID of pet that needs to be updated"
	// openapi:param x-agent-id header string true "Agent ID for the request"
	// openapi:success 204
	// openapi:failure 400 ErrorResponse
	GetPetByID(petId string) (*GetPetByIDResponse, error)
	SampleMethod3(param4 string, param5 string, param6 string) error
}

func main() {}
