package main

import (
	"ecommerce-service/graph"
	"ecommerce-service/utils"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

const defaultPort = "8081"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("🚨Error loading .env file")
	}

	utils.InitialiseDB()
	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())
	app.All("/graphql", GraphqlHandler)
	app.All("/query", QueryHandler)
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", defaultPort)
	log.Fatal(app.Listen(":" + defaultPort))
}

func QueryHandler(c *fiber.Ctx) error {
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))
	gqlHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.ServeHTTP(w, r)
	})
	fasthttpadaptor.NewFastHTTPHandler(gqlHandler)(c.Context())
	return nil
}

func GraphqlHandler(c *fiber.Ctx) error {
	playground := playground.Handler("GraphQL playground", "/query")
	fasthttpadaptor.NewFastHTTPHandler(playground)(c.Context())
	return nil
}
