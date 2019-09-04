package handlers

import (
	"context"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/machinebox/graphql"
)

// RegisterRequest is the request received to register a user.
type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// InsertUsersResponse is the response from the data service when
// the `insert_users` mutation completes.
type InsertUsersResponse struct {
	Users struct {
		AffectedRows int `json:"affected_rows"`
	} `json:"users"`
}

// Register a new user.
func Register(c echo.Context) error {
	var graphqlAdminSecret string
	graphqlAdminSecret = os.Getenv("RESPONSE_DATA_SECRET")

	params := RegisterRequest{}
	c.Bind(&params)

	client := c.Get("gql").(*graphql.Client)

	req := graphql.NewRequest(`
		mutation (
			$first_name: String!,
			$last_name: String!,
			$email: String!,
			$password: String!
		) {
			users: insert_users(
				objects: {
					first_name: $first_name,
					last_name: $last_name,
					email: $email,
					password: $password
				}
			) {
				affected_rows
			}
		}
	`)

	req.Header.Add("X-Hasura-Admin-Secret", graphqlAdminSecret)

	req.Var("first_name", params.FirstName)
	req.Var("last_name", params.LastName)
	req.Var("email", params.Email)
	req.Var("password", params.Password)

	ctx := context.Background()

	res := &InsertUsersResponse{}
	if err := client.Run(ctx, req, res); err != nil {
		return err
	}

	if res.Users.AffectedRows == 1 {
		return c.JSON(http.StatusOK, echo.Map{
			"success": true,
		})
	}

	return c.JSON(http.StatusInternalServerError, echo.Map{
		"error": echo.Map{
			"message": "There was a problem registering your account.",
		},
	})
}
