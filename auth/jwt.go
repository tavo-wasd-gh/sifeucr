package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Usuario string
	Cuenta  string
	jwt.RegisteredClaims
}

func JwtSet(w http.ResponseWriter, secure bool, name, usuario, cuenta string, expires time.Time, secret string) error {
	if name == "" || usuario == "" || cuenta == "" {
		return fmt.Errorf("missing claims")
	}

	claims := &Claims{
		usuario,
		cuenta,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Secure:   secure,
		Name:     name,
		Value:    token,
		Expires:  expires,
	})

	return nil
}

func JwtValidate(r *http.Request, name string, secret string) (string, string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", "", err
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.Usuario, claims.Cuenta, nil
	}

	return "", "", fmt.Errorf("invalid token or claims")
}
