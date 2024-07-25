package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

var Crypto *crypto

type CryptoType int

const (
	CryptoTypeDES CryptoType = iota //秘钥长度8字节 也就是64位
	CryptoTypeAES                   //秘钥长度位16 24 32 字节 也就是128 192 256位。
	CryptoType3DES
)

type crypto struct {
	base64 *base64.Encoding
}

func init() {
	Crypto = NewCrypto(base64.RawURLEncoding)
}

func NewCrypto(encoding *base64.Encoding) *crypto {
	return &crypto{base64: encoding}
}

// Encrypt DES加密
func (this *crypto) Encrypt(originalBytes, key []byte, scType CryptoType, ivs ...[]byte) ([]byte, error) {
	// 1、实例化密码器block(参数为密钥)
	var err error
	var block cipher.Block
	switch scType {
	case CryptoTypeDES:
		block, err = des.NewCipher(key)
	case CryptoType3DES:
		block, err = des.NewTripleDESCipher(key)
	case CryptoTypeAES:
		block, err = aes.NewCipher(key)
	default:
		block, err = nil, errors.New("CryptoType unknown")
	}
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	//fmt.Println("---blockSize---", blockSize)
	// 2、对明文填充字节(参数为原始字节切片和密码对象的区块个数)
	paddingBytes := PKCS7Padding(originalBytes, blockSize)
	//fmt.Println("填充后的字节切片：", paddingBytes)
	// 3、 实例化加密模式(参数为密码对象和密钥)
	var iv = key[:blockSize]
	if len(ivs) > 0 {
		iv = ivs[0]
	}
	blockMode := cipher.NewCBCEncrypter(block, iv)
	//fmt.Println("加密模式：", blockMode)
	// 4、对填充字节后的明文进行加密(参数为加密字节切片和填充字节切片)
	cipherBytes := make([]byte, len(paddingBytes))
	blockMode.CryptBlocks(cipherBytes, paddingBytes)
	return cipherBytes, nil
}

// Decrypt 解密字节切片，返回字节切片
func (this *crypto) Decrypt(cipherBytes, key []byte, scType CryptoType, ivs ...[]byte) ([]byte, error) {
	// 1、实例化密码器block(参数为密钥)
	var err error
	var block cipher.Block
	switch scType {
	case CryptoTypeDES:
		block, err = des.NewCipher(key)
	case CryptoType3DES:
		block, err = des.NewTripleDESCipher(key)
	case CryptoTypeAES:
		block, err = aes.NewCipher(key)
	default:
		block, err = nil, errors.New("CryptoType unknown")
	}
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	// 2、 实例化解密模式(参数为密码对象和密钥)
	var iv = key[:blockSize]
	if len(ivs) > 0 {
		iv = ivs[0]
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	// fmt.Println("解密模式：", blockMode)
	// 3、对密文进行解密(参数为加密字节切片和填充字节切片)
	paddingBytes := make([]byte, len(cipherBytes))
	blockMode.CryptBlocks(paddingBytes, cipherBytes)
	// 4、去除填充字节(参数为填充切片)
	originalBytes := PKCS7UnPadding(paddingBytes)
	return originalBytes, nil
}

func (this *crypto) DESEncrypt(original, secret string) (string, error) {
	chipperByte, err := this.Encrypt([]byte(original), []byte(secret), CryptoTypeDES)
	if err != nil {
		return "", err
	}
	base64str := this.base64.EncodeToString(chipperByte)
	return base64str, nil
}

func (this *crypto) DESDecrypt(chipper, secret string) (string, error) {
	base64Byte, err := this.base64.DecodeString(chipper)
	if err != nil {
		return "", err
	}
	chipperByte, err := this.Decrypt(base64Byte, []byte(secret), CryptoTypeDES)
	if err != nil {
		return "", err
	}
	return string(chipperByte), nil
}

func (this *crypto) AESEncrypt(original, secret string, ivs ...string) (string, error) {
	var chipperByte []byte
	var err error
	if len(ivs) > 0 {
		iv := []byte(ivs[0])
		chipperByte, err = this.Encrypt([]byte(original), []byte(secret), CryptoTypeAES, iv)
	} else {
		chipperByte, err = this.Encrypt([]byte(original), []byte(secret), CryptoTypeAES)
	}
	//chipperByte, err := this.Encrypt([]byte(original), []byte(secret), CryptoTypeAES)
	if err != nil {
		return "", err
	}
	base64str := this.base64.EncodeToString(chipperByte)
	return base64str, nil
}

func (this *crypto) AESDecrypt(chipper, secret string, ivs ...string) (string, error) {
	base64Byte, err := this.base64.DecodeString(chipper)
	if err != nil {
		return "", err
	}
	var chipperByte []byte
	if len(ivs) > 0 {
		iv := []byte(ivs[0])
		chipperByte, err = this.Decrypt(base64Byte, []byte(secret), CryptoTypeAES, iv)
	} else {
		chipperByte, err = this.Decrypt(base64Byte, []byte(secret), CryptoTypeAES)
	}
	if err != nil {
		return "", err
	}
	return string(chipperByte), nil
}

// GCMEncrypt
// AES-GCM（Galois/Counter Mode），因为它结合了AES的高强度加密和GCM的认证机制，能够提供很好的安全性并且生成的密文相对较短。
func (this *crypto) GCMEncrypt(text string, secret string, encode *base64.Encoding) (string, error) {
	//key := []byte("32-byte-long-key-here") // 应该是一个32字节的密钥
	key := []byte(secret)
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)
	if encode == nil {
		encode = base64.StdEncoding
	}

	return encode.EncodeToString(ciphertext), nil
}

func (this *crypto) GCMDecrypt(encryptedText string, secret string, encode *base64.Encoding) (string, error) {
	key := []byte(secret) // 应该是一个32字节的密钥
	if encode == nil {
		encode = base64.StdEncoding
	}
	ciphertext, err := encode.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", err
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// PKCS7Padding 填充字节的函数
func PKCS7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	//fmt.Println("要填充的字节：", padding)
	// 初始化一个元素为padding的切片
	slice1 := []byte{byte(padding)}
	slice2 := bytes.Repeat(slice1, padding)
	return append(data, slice2...)
}

// PKCS7UnPadding 去除填充字节的函数
func PKCS7UnPadding(data []byte) []byte {
	unpadding := data[len(data)-1]
	result := data[:(len(data) - int(unpadding))]
	return result
}

func MD5(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func SHA256(message string) string {
	hash := sha256.New()
	hash.Write([]byte(message))
	return hex.EncodeToString(hash.Sum(nil))
}
