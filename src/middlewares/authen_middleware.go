package middlewares

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"joodpakpaw/config"
	"joodpakpaw/internal/exception"
	"joodpakpaw/src/models"
)

// JwtAuthOptional is a middleware that optionally parses JWT token
// If token is valid, it sets authUser in context. If not, it continues without error.
func (m *Middleware) JwtAuthOptional() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// No token provided, continue without auth
			return c.Next()
		}

		// Extract token from "Bearer <token>"
		const prefix = "Bearer "
		if len(authHeader) <= len(prefix) {
			// Invalid format, continue without auth
			return c.Next()
		}
		tokenString := authHeader[len(prefix):]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Global.Auth.SecretKey), nil
		})
		if err != nil || !token.Valid {
			// Invalid token, continue without auth
			return c.Next()
		}

		// Extract claims and set user context
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Next()
		}

		if content, ok := claims["dat"].(map[string]interface{}); ok {
			if contentUser, ok := content["user"].(map[string]interface{}); ok {
				user := new(models.User)
				user.ID = int(contentUser["id"].(float64))
				user.Email = contentUser["email"].(string)

				if firstName, ok := contentUser["firstName"].(string); ok {
					user.FirstName = firstName
				}
				if lastName, ok := contentUser["lastName"].(string); ok {
					user.LastName = lastName
				}

				if roles, ok := contentUser["roles"].([]interface{}); ok {
					user.Roles = make([]models.Role, 0, len(roles))
					for _, r := range roles {
						if roleMap, ok := r.(map[string]interface{}); ok {
							role := models.Role{
								BaseModel: models.BaseModel{
									ID: int(roleMap["id"].(float64)),
								},
								Name: roleMap["name"].(string),
							}
							user.Roles = append(user.Roles, role)
						}
					}
				}

				c.Locals("authUser", user)
			}
		}

		return c.Next()
	}
}

func (m *Middleware) JwtAuthProtected() func(*fiber.Ctx) error {
	return jwtware.New(jwtware.Config{
		// Set the JWT signing key
		SigningKey: jwtware.SigningKey{Key: []byte(config.Global.Auth.SecretKey)},

		// Set lookup function
		TokenLookup: "header:Authorization",

		// Set auth scheme
		AuthScheme: "Bearer",

		// Set context key
		ContextKey: "auth",

		// Handle success case
		SuccessHandler: func(c *fiber.Ctx) error {
			// Get user from context
			token := c.Locals("auth").(*jwt.Token)

			// using claims data for get user from database
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				return exception.HttpErrorResponseMapping(c, fiber.StatusBadRequest, exception.InvalidTokenClaimResponseError, exception.ErrInvalidTokenClaim)
			}

			// Mapping content
			if content, ok := claims["dat"].(map[string]interface{}); ok {
				// Get user model
				if contentUser, ok := content["user"].(map[string]interface{}); ok {
					user := new(models.User)

					// Decode to struct
					user.ID = int(contentUser["id"].(float64))
					user.Email = contentUser["email"].(string)

					// Decode firstName and lastName if available
					if firstName, ok := contentUser["firstName"].(string); ok {
						user.FirstName = firstName
					}
					if lastName, ok := contentUser["lastName"].(string); ok {
						user.LastName = lastName
					}

					// Roles (many-to-many)
					if roles, ok := contentUser["roles"].([]interface{}); ok {
						user.Roles = make([]models.Role, 0, len(roles))
						for _, r := range roles {
							if roleMap, ok := r.(map[string]interface{}); ok {
								role := models.Role{
									BaseModel: models.BaseModel{
										ID: int(roleMap["id"].(float64)),
									},
									Name: roleMap["name"].(string),
								}
								user.Roles = append(user.Roles, role)
							}
						}
					}

					// Set auth user to context
					c.Locals("authUser", user)
				} else {
					return exception.HttpErrorResponseMapping(c, fiber.StatusBadRequest, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
				}
			} else {
				return exception.HttpErrorResponseMapping(c, fiber.StatusBadRequest, exception.InvalidTokenClaimDataResponseError, exception.ErrInvalidTokenClaimData)
			}

			return c.Next()
		},

		// Handle error cases
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Check if Authorization header is missing
			authHeader := c.Get("Authorization")
			if authHeader == "" {
				return exception.HttpErrorResponseMapping(c, fiber.StatusUnauthorized, exception.TokenMissingResponseError, exception.ErrTokenMissing)
			}

			// Check if token is expired
			if err.Error() == "token is expired" || err.Error() == "token has invalid claims: token is expired" {
				return exception.HttpErrorResponseMapping(c, fiber.StatusUnauthorized, exception.ExpiredAuthorizationTokenResponseError, exception.ErrExpiredAuthorizationToken)
			}

			// Other invalid token errors
			return exception.HttpErrorResponseMapping(c, fiber.StatusUnauthorized, exception.InvalidAuthorizationTokenResponseError, exception.ErrInvalidAuthorizationToken)
		},
	})
}
