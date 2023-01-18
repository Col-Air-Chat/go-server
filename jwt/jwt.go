package jwt

import (
	"col-air-go/util"
	"github.com/dgrijalva/jwt-go"
	"time"
)

var jwtKey = []byte("col-air-go-jwt-key#60A9AB6654EB706644FD6360A71328616AD7C142")

type Claims struct {
	UserId string
	jwt.StandardClaims
}

func ParseToken(tokenString string, subject string) (*Claims, error) {
	tokenDecrypt, err := util.Decrypt([]byte(tokenString))
	if err != nil {
		return nil, err
	}
	token, err := jwt.ParseWithClaims(string(tokenDecrypt), &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, err
	}
	if claims.Subject != subject {
		return nil, err
	}
	return claims, nil
}

func ClaimsCreate(userId string, expiresAt int64, subject string) *Claims {
	return &Claims{
		UserId: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
			IssuedAt:  time.Now().Unix(),
			Issuer:    "col-air-go-token-issuer",
			Subject:   subject,
		},
	}
}

func GenerateToken(userId string) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := ClaimsCreate(userId, expirationTime.Unix(), "col-air-go-token-main")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	tokenEnCrypt, err := util.Encrypt([]byte(tokenString))
	if err != nil {
		return "", err
	}
	return string(tokenEnCrypt), nil
}

func GenerateHeartbeatToken(userId string) (string, error) {
	expirationTime := time.Now().Add(60 * time.Second)
	claims := ClaimsCreate(userId, expirationTime.Unix(), "col-air-go-token-heartbeat")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	tokenEnCrypt, err := util.Encrypt([]byte(tokenString))
	if err != nil {
		return "", err
	}
	return string(tokenEnCrypt), nil
}

func RenewToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString, "col-air-go-token-main")
	if err != nil {
		return "", err
	}
	return GenerateToken(claims.UserId)
}

func RenewHeartbeatToken(tokenString string) (string, error) {
	Claims, err := ParseToken(tokenString, "col-air-go-token-heartbeat")
	if err != nil {
		return "", err
	}
	return GenerateHeartbeatToken(Claims.UserId)
}
