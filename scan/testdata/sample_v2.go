package main

import (
	"log"
	"net/http"
)

// @openapi:path /users
type userController struct{}

// @openapi:summary Retrieves a list of users.
// @openapi:description Returns a paginated list of users.
// @openapi:method GET
// @openapi:operationId getUsers
// @openapi:query limit int false "The number of items to return (max 100)"
// @openapi:query offset int false "The index of the first item to return"
// @openapi:response 200 []User "A list of users"
// @openapi:response 400 BadRequestResponse "Bad request"
// @openapi:response 500 InternalServerErrorResponse "Internal server error"
func (u *userController) getUsers(w http.ResponseWriter, r *http.Request) {
	// implementation here
}

// @openapi:path /users/{id}
type userDetailController struct{}

// @openapi:summary Retrieves a single user by ID.
// @openapi:description Returns the user with the specified ID.
// @openapi:method GET
// @openapi:operationId getUserById
// @openapi:param id int true "The ID of the user to retrieve"
// @openapi:response 200 User "The requested user"
// @openapi:response 400 BadRequestResponse "Bad request"
// @openapi:response 404 NotFoundResponse "User not found"
// @openapi:response 500 InternalServerErrorResponse "Internal server error"
func (u *userDetailController) getUserById(w http.ResponseWriter, r *http.Request) {
	// implementation here
}

func main() {

	// Serve the OpenAPI document on /openapi.json
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
