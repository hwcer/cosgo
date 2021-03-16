package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"encoding/base64"
	"errors"
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

//Encrypt DES加密
func (this *crypto) Encrypt(originalBytes, key []byte, scType CryptoType) ([]byte, error) {
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
	paddingBytes := PKCSSPadding(originalBytes, blockSize)
	//fmt.Println("填充后的字节切片：", paddingBytes)
	// 3、 实例化加密模式(参数为密码对象和密钥)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	//fmt.Println("加密模式：", blockMode)
	// 4、对填充字节后的明文进行加密(参数为加密字节切片和填充字节切片)
	cipherBytes := make([]byte, len(paddingBytes))
	blockMode.CryptBlocks(cipherBytes, paddingBytes)
	return cipherBytes, nil
}

// SCDecrypt 解密字节切片，返回字节切片
func (this *crypto) Decrypt(cipherBytes, key []byte, scType CryptoType) ([]byte, error) {
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
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	// fmt.Println("解密模式：", blockMode)
	// 3、对密文进行解密(参数为加密字节切片和填充字节切片)
	paddingBytes := make([]byte, len(cipherBytes))
	blockMode.CryptBlocks(paddingBytes, cipherBytes)
	// 4、去除填充字节(参数为填充切片)
	originalBytes := PKCSSUnPadding(paddingBytes)
	return originalBytes, nil
}

func (this *crypto) DESEncrypt(original, secret string) (string, error) {
	chipperByte, err := this.Encrypt([]byte(original), []byte(secret), CryptoTypeDES)
	if err != nil {
		return "", err
	}
	base64str := base64.URLEncoding.EncodeToString(chipperByte)
	return base64str, nil
}

func (this *crypto) DESDecrypt(chipper, secret string) (string, error) {
	base64Byte, err := base64.URLEncoding.DecodeString(chipper)
	if err != nil {
		return "", err
	}
	chipperByte, err := this.Decrypt(base64Byte, []byte(secret), CryptoTypeDES)
	if err != nil {
		return "", err
	}
	return string(chipperByte), nil
}

func (this *crypto) AESEncrypt(original, secret string) (string, error) {
	chipperByte, err := this.Encrypt([]byte(original), []byte(secret), CryptoTypeAES)
	if err != nil {
		return "", err
	}
	base64str := this.base64.EncodeToString(chipperByte)
	return base64str, nil
}

func (this *crypto) AESDecrypt(chipper, secret string) (string, error) {
	base64Byte, err := this.base64.DecodeString(chipper)
	if err != nil {
		return "", err
	}
	chipperByte, err := this.Decrypt(base64Byte, []byte(secret), CryptoTypeAES)
	if err != nil {
		return "", err
	}
	return string(chipperByte), nil
}

// PKCSSPadding 填充字节的函数
func PKCSSPadding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	//fmt.Println("要填充的字节：", padding)
	// 初始化一个元素为padding的切片
	slice1 := []byte{byte(padding)}
	slice2 := bytes.Repeat(slice1, padding)
	return append(data, slice2...)
}

// PKCSSUnPadding 去除填充字节的函数
func PKCSSUnPadding(data []byte) []byte {
	unpadding := data[len(data)-1]
	result := data[:(len(data) - int(unpadding))]
	return result
}
