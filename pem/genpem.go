package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

/*
用于生成公钥与私钥
*/
func main() {
	fmt.Println("开始运行...")
	prvKey, pubKey := GenRsaKey()
	fmt.Println(string(prvKey))
	fmt.Println(string(pubKey))
	err := os.WriteFile("prvKey.pem", prvKey, 0666)
	if err != nil {
		fmt.Println("生成失败", err)
	}
	err = os.WriteFile("pubKey.pem", pubKey, 0666)
	if err != nil {
		fmt.Println("生成失败", err)
	}
	fmt.Println("生成成功...")
}
func GenRsaKey() (prvkey, pubkey []byte) {
	// 生成私钥文件
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	prvkey = pem.EncodeToMemory(block)
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		panic(err)
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	pubkey = pem.EncodeToMemory(block)
	return
}
