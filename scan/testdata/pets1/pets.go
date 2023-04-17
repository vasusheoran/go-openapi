package main

import "encoding/json"

// Pet ...
// openapi:schema pet
type Pet struct {
	// openapi:description Name of the pet
	// openapi:example "rambo"
	// openapi:nullable
	// openapi:format text
	// openapi:default "tommy"
	Name string `json:"name"`
	// This is a sample field comment
	// openapi:description Type of pet
	// openapi:nullable
	// openapi:name category
	Category Category `json:"category"`
}

// Platform is enum to reflect platform type
// openapi:schema
type Platform string

// GetPets ...
// openapi:schema GetAllPets
type GetPets struct {
	// openapi:description Returns list of pets
	// openapi:format array
	// openapi:name pet
	Pets []Pet `json:"pets"`
	// openapi:name Platform
	Platform Platform `json:"platform"`
}

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

// PetsInterface This is a sample interface comment
// Interface are used to create tags. They must have `name` annotation associated with them.
type PetsInterface interface {
	// GetPets Fetches pets
	// openapi:operation GET /pets/ getPet
	// openapi:summary Fetches pet
	// openapi:description Fetches all pet
	// openapi:tag pets
	// openapi:consumes application/json
	// openapi:produces application/json
	// openapi:param x-agent-id header string true --- Agent ID for the request
	// openapi:response 200 GetAllPets --- Response for GetPetByID API
	// openapi:response 400 ErrorResponse --- Error
	GetPets(petId string) (GetPets, error)
}

func main() {}
