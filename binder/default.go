package binder

var Default Binder = Json

const (
	HeaderAccept      = "Accept"
	HeaderContentType = "Content-Type"
)

type ContentTypeMod int8

const (
	ContentTypeModReq ContentTypeMod = iota
	ContentTypeModRes
)

func GetContentType(meta map[string]string, mod ContentTypeMod) (r Binder) {
	var k string
	if mod == ContentTypeModReq {
		k = HeaderContentType
	} else {
		k = HeaderAccept
	}
	if ct := meta[k]; ct != "" {
		r = New(ct)
	}
	if r == nil && mod == ContentTypeModRes {
		r = GetContentType(meta, ContentTypeModReq) //保持和请求时一致
	}
	if r == nil {
		r = Default
	}
	return
}
