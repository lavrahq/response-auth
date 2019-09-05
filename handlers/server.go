package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/machinebox/graphql"
)

// ServersByCredentialsResponse is the expected response.
type ServersByCredentialsResponse struct {
	Servers []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"servers"`
}

// ServerCredentials are the credentials to use to login
type ServerCredentials struct {
	ID     string
	Secret string
}

// ResponseServerNamespacedClaims claims to be added under the Response namespace.
type ResponseServerNamespacedClaims struct {
	AllowedRoles []string `json:"x-hasura-allowed-roles"`
	DefaultRole  string   `json:"x-hasura-default-role"`
	ServerID     string   `json:"x-hasura-server-id"`
}

// JwtServerTokenClaims is the token claims to be added to the token.
type JwtServerTokenClaims struct {
	jwt.StandardClaims
	Name string `json:"name"`

	Response ResponseServerNamespacedClaims `json:"https://hasura.io/jwt/claims"`
}

func createServerToken(u *ServersByCredentialsResponse) (string, error) {
	roles := []string{"server"}

	// setup the claims
	claims := &JwtServerTokenClaims{
		StandardClaims: jwt.StandardClaims{
			Subject:  u.Servers[0].ID,
			IssuedAt: time.Now().Unix(),
			Issuer:   "response-auth.service",
		},

		Name: u.Servers[0].Name,
		Response: ResponseServerNamespacedClaims{
			AllowedRoles: roles,
			DefaultRole:  "server",
			ServerID:     u.Servers[0].ID,
		},
	}

	// create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(os.Getenv("RESPONSE_AUTH_JWT_SECRET")))
}

// ServerLogin authenticates a user by email/password using the data service.
func ServerLogin(c echo.Context) error {
	var graphqlAdminSecret string
	graphqlAdminSecret = os.Getenv("RESPONSE_DATA_SECRET")

	creds := ServerCredentials{}
	c.Bind(&creds)

	client := c.Get("gql").(*graphql.Client)

	req := graphql.NewRequest(`
		query ($id: uuid!, $secret: String!) {
			servers: servers_by_credentials(
				args: {
					server_id: $id,
					server_secret: $secret
				}
			) {
				id
				name
			}
		}
	`)

	req.Header.Add("X-Hasura-Admin-Secret", graphqlAdminSecret)

	req.Var("id", creds.ID)
	req.Var("secret", creds.Secret)

	ctx := context.Background()

	res := &ServersByCredentialsResponse{}
	if err := client.Run(ctx, req, res); err != nil {
		return err
	}

	if len(res.Servers) == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": map[string]interface{}{
				"message": "The credentials provided were not valid.",
			},
		})
	}

	if len(res.Servers) == 1 {
		token, err := createServerToken(res)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error": map[string]interface{}{
					"message": "There was a problem authenticating you. Please try again.",
				},
			})
		}

		return c.JSON(http.StatusAccepted, map[string]interface{}{
			"token": token,
		})
	}

	return c.JSON(http.StatusUnauthorized, map[string]interface{}{
		"error": map[string]interface{}{
			"message": "There was a problem authenticating you. Please try again.",
		},
	})
}
