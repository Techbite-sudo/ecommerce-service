package main

import (
	"context"
	"crypto/rand"
	"ecommerce-service/graph"
	"ecommerce-service/middleware"
	"ecommerce-service/utils"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Error loading .env file")
	}

	// Initialize OpenID Connect
	if err := middleware.InitAuth(); err != nil {
		log.Fatalf("Failed to initialize auth: %v", err)
	}

	// Initialize database
	utils.InitialiseDB()

	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	// Security Middleware
	app.Use(helmet.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: os.Getenv("ALLOWED_ORIGINS"),
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, OPTIONS",
	}))
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

	// Rate Limiting
	if os.Getenv("ENV") == "production" {
		app.Use(middleware.RateLimiter())
	}

	// Auth Routes
	authGroup := app.Group("/auth")
	authGroup.Get("/login", handleLogin)
	authGroup.Get("/callback", handleCallback)

	// GraphQL routes
	apiGroup := app.Group("/api")
	apiGroup.Use(middleware.AuthMiddleware())
	apiGroup.All("/query", QueryHandler)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// GraphQL playground (development only)
	if os.Getenv("ENV") != "production" {
		app.All("/graphql", GraphqlHandler)
		log.Printf("GraphQL Playground available at: http://localhost:%s/graphql", os.Getenv("PORT"))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	if os.Getenv("ENV") != "production" {
		return c.Status(code).JSON(fiber.Map{
			"error": message,
			"stack": err.Error(),
		})
	}

	return c.Status(code).JSON(fiber.Map{
		"error": message,
	})
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func handleLogin(c *fiber.Ctx) error {
	state := generateState()
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		Secure:   os.Getenv("ENV") == "production",
		HTTPOnly: true,
		SameSite: "Lax",
	})

	redirectURI := os.Getenv("FRONTEND_URL") + "/auth/callback"
	authURL := middleware.GetAuthCodeURL(state, redirectURI)
	return c.Redirect(authURL)
}

func handleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	// Verify state
	savedState := c.Cookies("oauth_state")
	if savedState != state {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid state",
		})
	}

	// Clear the state cookie
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
	})

	// Exchange code for token
	token, err := middleware.ExchangeCodeForToken(c.Context(), code)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"type":  "Bearer",
	})
}

func QueryHandler(c *fiber.Ctx) error {
	// Create new GraphQL server
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers:  &graph.Resolver{},
		Directives: graph.DirectiveRoot{},
	}))

	// Add error handling
	srv.SetErrorPresenter(func(ctx context.Context, e error) *gqlerror.Error {
		err := graphql.DefaultErrorPresenter(ctx, e)

		if os.Getenv("ENV") != "production" {
			// Include stack trace in development
			err.Extensions = map[string]interface{}{
				"stack": fmt.Sprintf("%+v", e),
			}
		}

		return err
	})

	// Create HTTP handler
	gqlHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add user from fiber context to request context
		user := c.Locals("user")
		ctx := context.WithValue(r.Context(), "user", user)
		r = r.WithContext(ctx)

		// Handle the request
		srv.ServeHTTP(w, r)
	})

	fasthttpadaptor.NewFastHTTPHandler(gqlHandler)(c.Context())
	return nil
}

func GraphqlHandler(c *fiber.Ctx) error {
	playground := playground.Handler("GraphQL playground", "/api/query")
	fasthttpadaptor.NewFastHTTPHandler(playground)(c.Context())
	return nil
}
