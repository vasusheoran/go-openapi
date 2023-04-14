package scan

import (
	"go/ast"
	"reflect"
	"testing"
)

func TestParseOpenAPIOperation(t *testing.T) {
	tests := []struct {
		name     string
		cg       *ast.CommentGroup
		comments []*ast.Comment
		want     *openAPIOperation
		wantErr  bool
	}{
		{
			name: "valid comments",
			cg: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// openapi:operation POST /pets2 createPet"},
					{Text: "// openapi:summary Adds a new pet to the store"},
					{Text: "// openapi:description Adds a new pet to the store"},
					{Text: "// openapi:tag pets2"},
					{Text: "// openapi:consumes application/json application/xml"},
					{Text: "// openapi:produces application/json application/xml"},
					{Text: "// openapi:body CreatePetRequest --- Request body to create Pets"},
					{Text: "// openapi:response 200 CreatePetResponse --- ResponseBody for CreatePet API"},
					{Text: "// openapi:response 204 --- OK"},
					{Text: "// openapi:response 400 ErrorResponse --- Error"},
					{Text: "// openapi:param name query string false --- Name of pet that needs to be updated"},
					{Text: "// openapi:param petId path string true --- ID of pet that needs to be updated"},
					{Text: "// openapi:param x-agent-id header string true --- Agent ID for the request"},
				},
			},
			want: &openAPIOperation{
				Method:      "POST",
				OperationID: "createPet",
				Path:        "/pets2",
				Summary:     "Adds a new pet to the store",
				Description: "Adds a new pet to the store",
				Tags:        []string{"pets2"},
				Consumes:    []string{"application/json", "application/xml"},
				Produces:    []string{"application/json", "application/xml"},
				Parameters: []*Parameter{
					{
						Name:        "name",
						In:          "query",
						Type:        "string",
						Required:    "false",
						Description: "Name of pet that needs to be updated",
					},
					{
						Name:        "petId",
						In:          "patj",
						Type:        "string",
						Required:    "true",
						Description: "ID of pet that needs to be updated",
					},
					{
						Name:        "x-agent-id",
						In:          "header",
						Type:        "string",
						Required:    "true",
						Description: "Name of pet that needs to be updated",
					},
				},
				RequestBody: &RequestBody{
					Name:        "CreatePetRequest",
					Description: "Request body to create Pets",
				},
				Responses: []*ResponseBody{
					{
						Name:        "CreatePetResponse",
						Description: "ResponseBody for CreatePet API",
						Code:        "200",
					},
					{
						Name:        "",
						Code:        "204",
						Description: "OK",
					},
					{
						Name:        "ErrorResponse",
						Description: "Error",
						Code:        "400",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid operation format",
			cg: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// openapi:operation POST"},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid body format",
			cg: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// openapi:body CreatePetRequest"},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid response format",
			cg: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// openapi:response 200"},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractOpenAPIOperation(tt.name, tt.cg)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractOpenAPIOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			//assert.Equal(t, tt.want, got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractOpenAPIOperation() got = %v, want %v", got, tt.want)
			}
		})
	}
}
