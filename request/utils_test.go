package request

import (
	"net/http/httptest"
	"testing"
)

// 签名串必须包含path与query:缺path时代理/中间人篡改路径后验签仍会通过
func TestAddressIncludesPathAndQuery(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com/v1/user/info?foo=1&bar=2", nil)
	if got := Address(req); got != "http://example.com/v1/user/info?foo=1&bar=2" {
		t.Errorf("Address 结果不匹配,实际%q", got)
	}

	//含需转义字符的path,EscapedPath保持percent-encoded形式
	req = httptest.NewRequest("GET", "http://example.com/", nil)
	req.URL.Path = "/a b/c/d"
	req.URL.RawPath = "/a%20b/c%2Fd"
	if got := Address(req); got != "http://example.com/a%20b/c%2Fd?" {
		t.Errorf("转义path结果不匹配,实际%q", got)
	}
}
