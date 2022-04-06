package registry

type RangeHandle func(name string, route *Service) bool

type Registry struct {
	*Options
	dict map[string]*Service
}

func New(opts *Options) *Registry {
	if opts == nil {
		opts = NewOptions()
	} else if opts.route == nil {
		opts.route = map[string]*Service{}
	}
	return &Registry{
		dict:    make(map[string]*Service),
		Options: opts,
	}
}

func (this *Registry) Len() int {
	return len(this.dict)
}

func (this *Registry) Get(name string) (srv *Service, ok bool) {
	name = this.Clean(name)
	srv, ok = this.dict[name]
	return
}

//Match 通过路径匹配Route,path必须是使用 Registry.Clean()处理后的
func (this *Registry) Match(path string) (srv *Service, ok bool) {
	//path = this.Clean(path)
	srv, ok = this.Options.route[path]
	return
}

func (this *Registry) Range(fn RangeHandle) {
	for k, r := range this.dict {
		if !fn(k, r) {
			return
		}
	}
	return
}

//Service GET OR CREATE
func (this *Registry) Service(name string) *Service {
	name = this.Clean(name)
	if r, ok := this.dict[name]; ok {
		return r
	}
	srv := NewService(name, this.Options)
	this.dict[srv.name] = srv
	return srv
}
