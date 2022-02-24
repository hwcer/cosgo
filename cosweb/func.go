package cosweb

import (
	"crypto/tls"
	"github.com/hwcer/cosgo/library/ioutil"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"unicode"
	"unicode/utf8"
)

func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

func strFirstToLower(str string) string {
	if len(str) < 1 {
		return ""
	}
	strArray := []rune(str)
	strArray[0] = unicode.ToLower(strArray[0])
	return string(strArray)
}

func TLSConfigTransform(key, pem string) (c *tls.Config, err error) {
	var bkey []byte
	if bkey, err = filepathOrContent(key); err != nil {
		return
	}

	var bcert []byte
	if bcert, err = filepathOrContent(pem); err != nil {
		return
	}

	c = new(tls.Config)
	c.Certificates = make([]tls.Certificate, 1)
	if c.Certificates[0], err = tls.X509KeyPair(bcert, bkey); err != nil {
		return
	}

	return
}

// createTLS starts an HTTPS server using certificates automatically installed from https://letsencrypt.org.
func TLSConfigAutocert() (c *tls.Config, err error) {
	autoTLSManager := autocert.Manager{Prompt: autocert.AcceptTOS}
	c = new(tls.Config)
	c.GetCertificate = autoTLSManager.GetCertificate
	c.NextProtos = append(c.NextProtos, acme.ALPNProto)
	return
}

func filepathOrContent(fileOrContent interface{}) (content []byte, err error) {
	switch v := fileOrContent.(type) {
	case string:
		return ioutil.ReadFile(v)
	case []byte:
		return v, nil
	default:
		return nil, ErrInvalidCertOrKeyType
	}
}

//通过文件或者证书内容获取TLSConfig
func TLSConfigParse(certFile, keyFile interface{}) (TLSConfig *tls.Config, err error) {
	var cert []byte
	if cert, err = filepathOrContent(certFile); err != nil {
		return
	}

	var key []byte
	if key, err = filepathOrContent(keyFile); err != nil {
		return
	}

	TLSConfig = new(tls.Config)
	TLSConfig.Certificates = make([]tls.Certificate, 1)
	if TLSConfig.Certificates[0], err = tls.X509KeyPair(cert, key); err != nil {
		return
	}
	return
}
