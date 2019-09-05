package main

import (
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/lavrahq/response-api/handlers"
	"github.com/machinebox/graphql"
)

// Echo is a the echo instance.
var Echo *echo.Echo

// Client is the GraphQL Client to the Response Data instance.
var Client *graphql.Client

func main() {

	var graphqlURL string

	graphqlURL = os.Getenv("RESPONSE_DATA_URL") + "/v1/graphql"

	if graphqlURL == "" {
		graphqlURL = "http://response-data:8080/v1/graphql"
	}

	Client = graphql.NewClient(graphqlURL)

	// Create an Echo instance.
	Echo = echo.New()
	Echo.Use(middleware.Logger())
	Echo.Use(middleware.Recover())
	Echo.Use(middleware.CORS())

	// Add the GraphQL Client to the Echo context.
	Echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("gql", Client)
			return next(c)
		}
	})

	// Setup Routes
	// Status Route
	Echo.GET("/", handlers.Index)

	// Current user information
	jwtMiddlewareConfig := middleware.JWTConfig{
		Claims:     &handlers.JwtTokenClaims{},
		SigningKey: []byte(os.Getenv("RESPONSE_AUTH_JWT_SECRET")),
	}
	Echo.GET("/user", handlers.User, middleware.JWTWithConfig(jwtMiddlewareConfig))

	// Register a user
	Echo.POST("/register", handlers.Register)

	// Login with credentials
	Echo.POST("/login", handlers.Login)

	// Login with server credentials
	Echo.POST("/server/login", handlers.ServerLogin)

	// Start the Echo Server
	Echo.Logger.Fatal((Echo.Start(":" + os.Getenv("PORT"))))
}
