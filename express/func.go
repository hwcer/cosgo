package express

import (
	"crypto/tls"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"io/ioutil"
	"reflect"
	"runtime"
)

func handlerName(h HandlerFunc) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
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
