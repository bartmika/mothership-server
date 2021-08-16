package utils

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// Generate the `access token` and `refresh token` for the secret key.
func GenerateJWTTokenPair(hmacSecret []byte, sessionUuid string, d time.Duration) (string, string, error) {
	//
	// Generate token.
	//

	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(d).Unix(),
		Issuer:    sessionUuid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(hmacSecret)
	if err != nil {
		return "", "", err
	}

	//
	// Generate refresh token.
	//

	claims = &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(d + time.Hour*24*7).Unix(), // Plus 7 days.
		Issuer:    sessionUuid,
	}

	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshTokenString, err := token.SignedString(hmacSecret)

	// Return our tokens
	return tokenString, refreshTokenString, err
}

// Validates either the `access token` or `refresh token` and returns either the
// `uuid` if success or error on failure.
func ProcessBearerToken(hmacSecret []byte, tokenString string) (string, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return hmacSecret, nil
	})

	if err == nil && token.Valid {
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			issuer := claims["iss"].(string)
			// m["exp"] := string(claims["exp"].(float64))
			return issuer, nil
		} else {
			return "", err
		}
	} else {
		return "", err
	}
	return "", nil
}
