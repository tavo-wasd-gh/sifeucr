package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID string
	jwt.RegisteredClaims
}

// jwtSet(w, true, "token", "id", time.Now().Add(15 * time.Minute), []byte("secret"))
func jwtSet(w http.ResponseWriter, secure bool, name, id_cuenta string, expires time.Time, secret []byte) error {
	claims := &Claims{
		id_cuenta,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
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

// jwtValidate(r, "token", []byte("secret"))
func jwtValidate(r *http.Request, name string, secret []byte) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.ID, nil
	}

	return "", fmt.Errorf("invalid token or claims")
}
