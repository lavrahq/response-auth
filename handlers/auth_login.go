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

// UsersByCredentialsResponse is the expected response.
type UsersByCredentialsResponse struct {
	Users []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Roles []struct {
			ID  string `json:"id"`
			Key string `json:"key"`
		} `json:"roles"`
	} `json:"users"`
}

// Credentials are the credentials to use to login
type Credentials struct {
	Email    string
	Password string
}

// ResponseNamespacedClaims claims to be added under the Response namespace.
type ResponseNamespacedClaims struct {
	AllowedRoles []string `json:"x-hasura-allowed-roles"`
	DefaultRole  string   `json:"x-hasura-default-role"`
	UserID       string   `json:"x-hasura-user-id"`
}

// JwtTokenClaims is the token claims to be added to the token.
type JwtTokenClaims struct {
	jwt.StandardClaims
	Name string `json:"name"`

	Response ResponseNamespacedClaims `json:"https://hasura.io/jwt/claims"`
}

func createToken(u *UsersByCredentialsResponse) (string, error) {
	roles := []string{"user"}

	for _, v := range u.Users[0].Roles {
		roles = append(roles, v.Key)
	}

	// setup the claims
	claims := &JwtTokenClaims{
		StandardClaims: jwt.StandardClaims{
			Subject:   u.Users[0].ID,
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Issuer:    "response-auth.service",
		},

		Name: u.Users[0].Name,
		Response: ResponseNamespacedClaims{
			AllowedRoles: roles,
			DefaultRole:  "user",
			UserID:       u.Users[0].ID,
		},
	}

	// create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(os.Getenv("RESPONSE_AUTH_JWT_SECRET")))
}

// AuthLogin returns the login route.
func AuthLogin(c echo.Context) error {
	var graphqlAdminSecret string
	graphqlAdminSecret = os.Getenv("RESPONSE_DATA_SECRET")

	creds := Credentials{}
	c.Bind(&creds)

	client := c.Get("gql").(*graphql.Client)

	req := graphql.NewRequest(`
		query ($email: String!, $password: String!) {
			users: users_by_credentials(
				args: {
					user_email: $email,
					user_password: $password
				}
			) {
				id
				roles {
					role {
						id
						key
					}
				}
			}
		}
	`)

	req.Header.Add("X-Hasura-Admin-Secret", graphqlAdminSecret)

	req.Var("email", creds.Email)
	req.Var("password", creds.Password)

	ctx := context.Background()

	res := &UsersByCredentialsResponse{}
	if err := client.Run(ctx, req, res); err != nil {
		return err
	}

	if len(res.Users) == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": map[string]interface{}{
				"message": "The credentials provided were not valid.",
			},
		})
	}

	if len(res.Users) == 1 {
		token, err := createToken(res)
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
