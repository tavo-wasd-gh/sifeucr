package main

import (
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("1234")

type jwtCustomClaims struct {
	IDCuenta string `json:"id_cuenta" db:"id_cuenta"`
	jwt.RegisteredClaims
}
