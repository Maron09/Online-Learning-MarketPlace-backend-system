package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/sikozonpc/ecom/db"
	"github.com/sikozonpc/ecom/types"
	"github.com/sikozonpc/ecom/utils"
)

type contextKey string

const UserKey contextKey = "user_id"
const RoleKey contextKey = "userRole"

func CreateJWT(secret []byte, userID int, role types.UserRole) (string, error) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	ext := db.PostgresConfig{
		JWTExpiration: db.ParseJWTExp(os.Getenv("JWT_EXPIRATION")),
	}

	expiration := time.Second * time.Duration(ext.JWTExpiration)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": strconv.Itoa(userID),
		"role":    role,
        "exp":    time.Now().Add(expiration).Unix(),
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}


// Decorator to wrap endpoints with custom authentication and role-based restriction
func WithJWTAuth(handlerFunc http.HandlerFunc, store types.UserStore, allowedRoles []types.UserRole) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Get the token from the user request
        tokenString := GetJwtTokenFromRequest(r)
        log.Printf("Extracted token: %s", tokenString)

        // Validate the JWT
        token, err := ValidateToken(tokenString)
        if err != nil {
            log.Printf("Failed to validate token: %v", err)
            PermissionDenied(w)
            return
        }
        log.Println("Token validated successfully")

        if !token.Valid {
            log.Println("Invalid token")
            PermissionDenied(w)
            return
        }

        // Extract claims and fetch the userID from the token
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            log.Println("Failed to extract claims from token")
            PermissionDenied(w)
            return
        }

        // Validate user_id claim
        var userID int
        switch id := claims["user_id"].(type) {
        case string:
            userID, err = strconv.Atoi(id)
            if err != nil {
                log.Printf("Failed to convert user_id to integer: %v", err)
                PermissionDenied(w)
                return
            }
        case float64: // Handling if user_id is stored as a number in the token
            userID = int(id)
        default:
            log.Println("user_id claim missing or invalid type in the token")
            PermissionDenied(w)
            return
        }

        // Validate role claim
        userRoleClaim, ok := claims["role"].(float64)
        if !ok {
            log.Println("role claim missing or invalid type in the token")
            PermissionDenied(w)
            return
        }

        // Convert the userRoleClaim to UserRole type
        userRole := types.UserRole(int(userRoleClaim))

        // Fetch the user from the database
        u, err := store.GetUserByID(userID)
        if err != nil {
            log.Printf("Failed to get user by ID: %v", err)
            PermissionDenied(w)
            return
        }

        // Check if the user's role is permitted
        roleAllowed := false
        for _, role := range allowedRoles {
            if u.Role == role || userRole == role { // Check both u.Role and userRole
                roleAllowed = true
                break
            }
        }

        if !roleAllowed {
            log.Printf("User with ID %d has role %d which is not permitted", userID, u.Role)
            PermissionDenied(w)
            return
        }

        // Set user ID and Role in the request context for future use in the handler
        ctx := context.WithValue(r.Context(), UserKey, u.ID)
        ctx = context.WithValue(ctx, RoleKey, u.Role)
        r = r.WithContext(ctx)

        // Call the original handler
        handlerFunc(w, r)
    }
}






func GetJwtTokenFromRequest(req *http.Request) string {
	authHeader := req.Header.Get("Authorization")
	if len(authHeader) > 7 && strings.ToLower(authHeader[0:6]) == "bearer" {
		return authHeader[7:]
	}
	return ""
}

func ValidateToken(t string) (*jwt.Token, error) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	secret := os.Getenv("JWTSecret")

	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		// Check that the token is signed using HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return the secret for signing and verification
		return []byte(secret), nil
	})

	return token, err
}


func PermissionDenied(writer http.ResponseWriter){
	utils.WriteError(writer, http.StatusForbidden, fmt.Errorf("permission denied"))
}


func GetUserIDFromContext(ctx context.Context) int {
	userID, ok := ctx.Value(UserKey).(int)

	if !ok {
		return -1
	}

	return userID
}