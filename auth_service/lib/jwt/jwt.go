package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/silver-eva/auth_service/auth_service/models"
)

var secretKey = []byte("supersecretkey")

// GenerateJWT generates a new JWT token.
func GenerateJWT(user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Name,
		"role":     user.Role,
		"password": user.Password,
		"exp":      time.Now().Add(time.Hour * 1).Unix(),
		"id":       user.Id,
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// DecodeJWT decodes and validates the JWT token.
func DecodeJWT(tokenStr string) (*models.DecodedToken, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		exp := int64(claims["exp"].(float64))
		return &models.DecodedToken{
			UserName: claims["username"].(string),
			UserRole: claims["role"].(string),
			Expired:  time.Unix(exp, 0),
			UserPass: claims["password"].(string),
			UserId:   claims["id"].(string),
		}, nil
	}

	return nil, errors.New("invalid token")
}
