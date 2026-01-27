package binder

var Default Binder = Json

const (
	HeaderAccept      = "Accept"
	HeaderContentType = "Content-Type"
)

func GetBinder(meta map[string]string, contentType ...string) (b Binder) {
	if len(contentType) == 0 {
		contentType = []string{HeaderContentType, HeaderAccept}
	}
	for _, m := range contentType {
		if v, ok := meta[m]; ok {
			if b = Get(v); b != nil {
				return
			}
		}
	}
	return Default
}
