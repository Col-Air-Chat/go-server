package jwt

import (
	"col-air-go/util"
	"github.com/dgrijalva/jwt-go"
	"time"
)

/**
 * TODO: RSA is too slow, so use HS256, but it is not safe
 * TODO: maybe find a better way to encrypt and decrypt
 */

var jwtKey = []byte("col-air-go-jwt-key#60A9AB6654EB706644FD6360A71328616AD7C142")

type Claims struct {
	UserId string
	jwt.StandardClaims
}

func ParseToken(tokenString string, subject string) (*Claims, error) {
	// resultChan := make(chan string)
	// go DecryptChannel(tokenString, resultChan)
	// tokenDecrypt := <-resultChan
	// if tokenDecrypt == "" {
	// 	return nil, error(nil)
	// }
	token, err := jwt.ParseWithClaims(string(tokenString), &Claims{}, func(token *jwt.Token) (interface{}, error) {
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

func ParseHeartbeatToken(tokenString string, subject string, mainToken string) (*Claims, error) {
	// resultChan := make(chan string)
	// go DecryptChannel(tokenString, resultChan)
	// tokenDecrypt := <-resultChan
	// if tokenDecrypt == "" {
	// 	return nil, error(nil)
	// }
	token, err := jwt.ParseWithClaims(string(tokenString), &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(string(jwtKey) + mainToken), nil
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
	// resultChan := make(chan string)
	// go EncryptChannel(tokenString, resultChan)
	// tokenEnCrypt := <-resultChan
	// if tokenEnCrypt == "" {
	// 	return "", err
	// }
	return string(tokenString), nil
}

func GenerateHeartbeatToken(userId string, mainToken string) (string, error) {
	expirationTime := time.Now().Add(60 * time.Second)
	mainClaims, err := ParseToken(mainToken, "col-air-go-token-main")
	if err != nil {
		return "", err
	}
	if mainClaims.UserId != userId {
		return "", err
	}
	claims := ClaimsCreate(userId, expirationTime.Unix(), "col-air-go-token-heartbeat")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(string(jwtKey) + mainToken))
	if err != nil {
		return "", err
	}
	// resultChan := make(chan string)
	// go EncryptChannel(tokenString, resultChan)
	// tokenEnCrypt := <-resultChan
	// if tokenEnCrypt == "" {
	// 	return "", err
	// }
	return string(tokenString), nil
}

func RenewToken(tokenString string) (string, string, error) {
	claims, err := ParseToken(tokenString, "col-air-go-token-main")
	if err != nil {
		return "", "", err
	}
	token, err := GenerateToken(claims.UserId)
	return token, claims.UserId, err
}

func RenewHeartbeatToken(tokenString string, mainToken string) (string, string, error) {
	Claims, err := ParseHeartbeatToken(tokenString, "col-air-go-token-heartbeat", mainToken)
	if err != nil {
		return "", "", err
	}
	token, err := GenerateHeartbeatToken(Claims.UserId, mainToken)
	return token, Claims.UserId, err
}

func EncryptChannel(value string, resultChan chan string) {
	tokenEnCrypt, err := util.Encrypt([]byte(value))
	if err != nil {
		resultChan <- err.Error()
	}
	resultChan <- string(tokenEnCrypt)
}

func DecryptChannel(value string, resultChan chan string) {
	tokenDecrypt, err := util.Decrypt([]byte(value))
	if err != nil {
		resultChan <- ""
	}
	resultChan <- string(tokenDecrypt)
}
