package express

import "net/http"

var RESTfulMethods = [...]string{
	http.MethodGet,
	http.MethodPut,
	http.MethodPost,
	http.MethodDelete,
}

type RESTful interface {
	GET(*Context) error    //用来获取资源
	PUT(*Context) error    //PUT用来更新资源
	POST(*Context) error   //用来新建资源（也可以用于更新资源）
	DELETE(*Context) error //用来删除资源
}
