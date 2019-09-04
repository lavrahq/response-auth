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

	// Add the GraphQL Client to the Echo context.
	Echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("gql", Client)
			return next(c)
		}
	})

	// Setup Routes
	Echo.GET("/", handlers.Index)

	auth := Echo.Group("/auth")
	auth.POST("/login", handlers.AuthLogin)

	// Start the Echo Server
	Echo.Logger.Fatal((Echo.Start(":8090")))
}
