package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("1234")

type jwtClaims struct {
	IDCuenta  string `json:"id_cuenta"`
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
		Name:    name,
		Value:   token,
		Expires: expires,
	})

	return nil
}

func jwtValidate(r *http.Request, name string) error {
	cookie, err := r.Cookie(name)
	if err != nil {
		return err
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return err
	} 

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}
