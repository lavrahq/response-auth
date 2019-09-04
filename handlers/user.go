package handlers

import (
	"context"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/machinebox/graphql"
)

// UsersQueryResponse is the response returned from the data service query.
type UsersQueryResponse struct {
	Users []struct {
		ID        string `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Roles     []struct {
			Role struct {
				ID          string `json:"id"`
				Key         string `json:"key"`
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"role"`
		} `json:"roles"`
	} `json:"users"`
}

// User returns the currently authenticated user's information directly from
// the data service.
func User(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*JwtTokenClaims)

	var graphqlAdminSecret string
	graphqlAdminSecret = os.Getenv("RESPONSE_DATA_SECRET")

	params := RegisterRequest{}
	c.Bind(&params)

	client := c.Get("gql").(*graphql.Client)

	req := graphql.NewRequest(`
		query ($user_id: uuid!) {
			users(
				where: {
					id: {
						_eq: $user_id
					}
				}
			) {
				id
				first_name
				last_name
				email
				roles {
					role {
						id
						key
						title
						description
					}
				}
			}
		}
	`)

	req.Header.Add("X-Hasura-Admin-Secret", graphqlAdminSecret)

	req.Var("user_id", claims.Response.UserID)

	ctx := context.Background()

	res := &UsersQueryResponse{}
	if err := client.Run(ctx, req, res); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res.Users[0])
}
