package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
)

// RSAKeyPair RSA密钥对
type RSAKeyPair struct {
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

// GenerateRSAKeyPair 生成RSA密钥对
func GenerateRSAKeyPair() (*RSAKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("生成RSA密钥对失败: %w", err)
	}

	return &RSAKeyPair{
		PublicKey:  &privateKey.PublicKey,
		PrivateKey: privateKey,
	}, nil
}

// EncryptWithPublicKey 使用公钥加密
func EncryptWithPublicKey(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
	if err != nil {
		return nil, fmt.Errorf("加密失败: %w", err)
	}
	return encrypted, nil
}

// DecryptWithPrivateKey 使用私钥解密
func DecryptWithPrivateKey(privateKey *rsa.PrivateKey, encrypted []byte) ([]byte, error) {
	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encrypted)
	if err != nil {
		return nil, fmt.Errorf("解密失败: %w", err)
	}
	return decrypted, nil
}

// PublicKeyToBase64 将公钥转换为Base64字符串
func PublicKeyToBase64(publicKey *rsa.PublicKey) string {
	publicKeyBytes := x509.MarshalPKCS1PublicKey(publicKey)
	return base64.StdEncoding.EncodeToString(publicKeyBytes)
}

// PrivateKeyToBase64 将私钥转换为Base64字符串
func PrivateKeyToBase64(privateKey *rsa.PrivateKey) string {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	return base64.StdEncoding.EncodeToString(privateKeyBytes)
}

// LoadPublicKeyFromBase64 从Base64加载公钥
func LoadPublicKeyFromBase64(base64Str string) (*rsa.PublicKey, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, fmt.Errorf("解码公钥失败: %w", err)
	}

	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %w", err)
	}

	return publicKey, nil
}

// LoadPrivateKeyFromBase64 从Base64加载私钥
func LoadPrivateKeyFromBase64(base64Str string) (*rsa.PrivateKey, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, fmt.Errorf("解码私钥失败: %w", err)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	return privateKey, nil
}

// LoadPublicKeyFromPEM 从PEM文件加载公钥
func LoadPublicKeyFromPEM(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取公钥文件失败: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("解码PEM失败")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %w", err)
	}

	return publicKey, nil
}

// LoadPrivateKeyFromPEM 从PEM文件加载私钥
func LoadPrivateKeyFromPEM(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("解码PEM失败")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	return privateKey, nil
}

// SavePublicKeyToPEM 保存公钥到PEM文件
func SavePublicKeyToPEM(publicKey *rsa.PublicKey, path string) error {
	publicKeyBytes := x509.MarshalPKCS1PublicKey(publicKey)
	block := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建公钥文件失败: %w", err)
	}
	defer file.Close()

	if err := pem.Encode(file, block); err != nil {
		return fmt.Errorf("写入公钥失败: %w", err)
	}

	return nil
}

// SavePrivateKeyToPEM 保存私钥到PEM文件
func SavePrivateKeyToPEM(privateKey *rsa.PrivateKey, path string) error {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建私钥文件失败: %w", err)
	}
	defer file.Close()

	if err := pem.Encode(file, block); err != nil {
		return fmt.Errorf("写入私钥失败: %w", err)
	}

	return nil
}
