package registry

type Registry struct {
	dict   map[string]*Service
	router *Router
}

func New() *Registry {
	return &Registry{
		dict:   make(map[string]*Service),
		router: NewRouter(),
	}
}

func (this *Registry) Len() int {
	return len(this.dict)
}

func (this *Registry) Get(name string) (srv *Service, ok bool) {
	prefix := Join(name)
	srv, ok = this.dict[prefix]
	return
}
func (this *Registry) Has(name string) (ok bool) {
	prefix := Join(name)
	_, ok = this.dict[prefix]
	return
}

//	func (this *Registry) Merge(r *Registry) (err error) {
//		this.Range(func(s *Service) bool {
//			prefix := s.prefix
//			if _, ok := this.dict[prefix]; !ok {
//				this.dict[prefix] = NewService(prefix, this.router)
//			}
//			if err = this.dict[prefix].Merge(s); err != nil {
//				return false
//			}
//			return true
//		})
//		return
//	}
//type ssr map[string]*Node

//	func (this *Registry) Reload(nodes map[string]*Node) (err error) {
//		ssc := make(map[string]ssr)
//		for _, v := range nodes {
//			sr := ssc[v.Service.Name()]
//			if sr == nil {
//				sr = make(ssr)
//				ssc[v.Service.Name()] = sr
//			}
//			sr[v.Name()] = v
//		}
//		for k, v := range ssc {
//			service, ok := this.Get(k)
//			if !ok {
//				return fmt.Errorf("service not found:%v", k)
//			}
//			if err = service.Reload(v); err != nil {
//				return err
//			}
//		}
//
//		return
//	}

//	func (this *Registry) Method() Method {
//		return this.router.Method()
//	}
func (this *Registry) Router() *Router {
	return this.router
}

// Search 通过路径匹配Route,path必须是使用 Registry.Clean()处理后的
func (this *Registry) Search(method string, paths ...string) []*Node {
	return this.router.Search(method, paths...)
}

func (this *Registry) Match(method string, paths ...string) (*Node, bool) {
	nodes := this.router.Search(method, paths...)
	if len(nodes) == 0 {
		return nil, false
	} else {
		return nodes[0], true
	}
}

// Service GET OR CREATE
func (this *Registry) Service(name string, hs ...Handler) *Service {
	prefix := Route(name)
	if r, ok := this.dict[prefix]; ok {
		return r
	}
	srv := NewService(prefix, this.router, hs...)
	this.dict[prefix] = srv
	return srv
}

// Range 遍历所有服务
func (this *Registry) Range(f func(service *Service) bool) {
	for _, srv := range this.dict {
		if !f(srv) {
			return
		}
	}
}

// Nodes 遍历所有节点
func (this *Registry) Nodes(f func(node *Node) bool) {
	for _, service := range this.dict {
		nodes := service.nodes
		for _, node := range nodes {
			if !f(node) {
				return
			}
		}
	}
}
