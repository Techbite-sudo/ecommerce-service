package middleware

import (
	"context"
	"ecommerce-service/models"
	"ecommerce-service/utils"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
)

var (
	provider     *oidc.Provider
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
)

type Claims struct {
	Email string `json:"email"`
	Sub   string `json:"sub"`
	Name  string `json:"name"`
}

func InitAuth() error {
	ctx := context.Background()
	var err error

	// Initialize OIDC provider
	provider, err = oidc.NewProvider(ctx, os.Getenv("OIDC_PROVIDER_URL"))
	if err != nil {
		return fmt.Errorf("failed to initialize OIDC provider: %v", err)
	}

	oauth2Config = oauth2.Config{
		ClientID:     os.Getenv("OIDC_CLIENT_ID"),
		ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OIDC_REDIRECT_URL"),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier = provider.Verifier(&oidc.Config{
		ClientID: os.Getenv("OIDC_CLIENT_ID"),
	})

	return nil
}

// AuthMiddleware handles authentication using OpenID Connect
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization format",
			})
		}

		// Verify the ID token
		ctx := context.Background()
		idToken, err := verifier.Verify(ctx, parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Extract claims
		var claims Claims
		if err := idToken.Claims(&claims); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to parse token claims",
			})
		}

		// Get or create user
		user, err := getOrCreateUser(claims)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "User processing failed",
			})
		}

		// Add user to context
		c.Locals("user", user)
		return c.Next()
	}
}

// Helper function to get or create user based on OIDC claims
func getOrCreateUser(claims Claims) (*models.User, error) {
	var user models.User

	// Try to find existing user
	err := utils.DB.Where("email = ?", claims.Email).First(&user).Error
	if err == nil {
		return &user, nil
	}

	// Create new user if not found
	user = models.User{
		Email: claims.Email,
		Names: claims.Name,
		Role:  models.RoleUser, // Default to regular user role
	}

	if err := utils.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// RequireAuth for GraphQL resolvers
func RequireAuth(ctx context.Context) (*models.User, error) {
	user := ctx.Value("user")
	if user == nil {
		return nil, errors.New("unauthorized")
	}
	return user.(*models.User), nil
}

// RequireRole for GraphQL resolvers
func RequireRole(ctx context.Context, roles ...models.Role) error {
	user, err := RequireAuth(ctx)
	if err != nil {
		return err
	}

	for _, role := range roles {
		if user.Role == role {
			return nil
		}
	}

	return errors.New("insufficient permissions")
}
