package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("1234")

type jwtClaims struct {
	IDCuenta string `json:"id_cuenta"`
	jwt.RegisteredClaims
}

// jwtSet(w, "jwt_token", "ASO-123", time.Now().Add(15 * time.Minute))
func jwtSet(w http.ResponseWriter, name, id_cuenta string, expires time.Time) error {
	claims := &jwtClaims{
		id_cuenta,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Secure:   ProductionEnvironment,
		Name:     name,
		Value:    token,
		Expires:  expires,
	})

	return nil
}

// jwtValidate(r, "jwt_token")
func jwtValidate(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		return claims.IDCuenta, nil
	}

	return "", fmt.Errorf("invalid token or claims")
}
