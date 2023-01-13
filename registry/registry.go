package registry

type Registry struct {
	dict   map[string]*Service
	router *Router
}

func New(router *Router) *Registry {
	if router == nil {
		router = NewRouter()
	}
	return &Registry{
		dict:   make(map[string]*Service),
		router: router,
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

func (this *Registry) Merge(r *Registry) (err error) {
	this.Range(func(s *Service) bool {
		prefix := s.prefix
		if _, ok := this.dict[prefix]; !ok {
			this.dict[prefix] = NewService(prefix, this.router)
		}
		if err = this.dict[prefix].Merge(s); err != nil {
			return false
		}
		return true
	})
	return
}

func (this *Registry) Router() *Router {
	return this.router
}

// Match 通过路径匹配Route,path必须是使用 Registry.Clean()处理后的
func (this *Registry) Match(paths ...string) (node *Node, ok bool) {
	nodes := this.router.Match(paths...)
	for i := len(nodes) - 1; i >= 0; i-- {
		if node, ok = nodes[i].Handle().(*Node); ok {
			return
		}
	}
	return
}

// Service GET OR CREATE
func (this *Registry) Service(name string) *Service {
	prefix := Join(name)
	if r, ok := this.dict[prefix]; ok {
		return r
	}
	srv := NewService(prefix, this.router)
	this.dict[prefix] = srv
	return srv
}

// Services 获取所有ServicePath
//func (this *Registry) Services() (r []*Service) {
//	for _, s := range this.dict {
//		r = append(r, s)
//	}
//	return
//}

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
		for _, node := range service.nodes {
			if !f(node) {
				return
			}
		}
	}
}
