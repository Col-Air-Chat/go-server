package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
)

var (
	privateKey *rsa.PrivateKey = GetPrivateKey("private.pem")
	publicKey  *rsa.PublicKey  = GetPublicKey("public.pem")
)

func GetPrivateKey(path string) *rsa.PrivateKey {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	info, _ := file.Stat()
	buf := make([]byte, info.Size())
	file.Read(buf)
	block, _ := pem.Decode(buf)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil
	}
	return privateKey
}

func GetPublicKey(path string) *rsa.PublicKey {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	info, _ := file.Stat()
	buf := make([]byte, info.Size())
	file.Read(buf)
	block, _ := pem.Decode(buf)
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil
	}
	publicKey := publicKeyInterface.(*rsa.PublicKey)
	return publicKey
}

func GenerateRSAKey(bits int) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return
	}
	X509PrivateKey := x509.MarshalPKCS1PrivateKey(privateKey)
	privateFile, err := os.Create("private.pem")
	if err != nil {
		return
	}
	defer privateFile.Close()
	privateBlock := pem.Block{Type: "RSA Private Key", Bytes: X509PrivateKey}
	pem.Encode(privateFile, &privateBlock)
	publicKey := privateKey.PublicKey
	X509PublicKey, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return
	}
	publicFile, err := os.Create("public.pem")
	if err != nil {
		return
	}
	defer publicFile.Close()
	publicBlock := pem.Block{Type: "RSA Public Key", Bytes: X509PublicKey}
	pem.Encode(publicFile, &publicBlock)
}

func Encrypt(plainText []byte) ([]byte, error) {
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, plainText)
	if err != nil {
		return nil, err
	}
	cipherText = Base64Encode(cipherText)
	return cipherText, err
}

func Decrypt(cipherText []byte) ([]byte, error) {
	cipherText, err := Base64Decode(cipherText)
	if err != nil {
		return nil, err
	}
	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherText)
	if err != nil {
		return nil, err
	}
	return plainText, err
}

func Base64Encode(src []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(src))
}

func Base64Decode(src []byte) ([]byte, error) {
	return base64.StdEncoding.DecodeString(string(src))
}
