package registry

type RangeHandle func(name string, route *Route) error

type Registry struct {
	*Options
	dict map[string]*Route
}

func New(opts *Options) *Registry {
	if opts == nil {
		opts = &Options{}
	}
	return &Registry{
		Options: opts,
		dict:    make(map[string]*Route),
	}
}

func (this *Registry) Route(name string) *Route {
	route := NewRoute(this.Options, name)
	if r, ok := this.dict[route.name]; ok {
		return r
	}
	this.dict[route.name] = route
	return route
}

func (this *Registry) Paths() (r []string) {
	for k, _ := range this.dict {
		r = append(r, k)
	}
	return
}

func (this *Registry) Range(fn RangeHandle) (err error) {
	for k, r := range this.dict {
		if err = fn(k, r); err != nil {
			return
		}
	}
	return
}
