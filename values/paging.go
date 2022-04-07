package values

type Paging struct {
	Rows   interface{} `json:"rows"`
	Page   int         `json:"page"`   //当前页
	Size   int         `json:"size"`   //每页大小
	Total  int         `json:"total"`  //总记录数
	Update int64       `json:"update"` //最后更新时间
}

func (this *Paging) Init(size int) {
	if this.Page == 0 {
		this.Page = 1
	}
	if this.Size == 0 {
		this.Size = size
	} else if this.Size > size {
		this.Size = size
	}
	this.Rows = make([]interface{}, 0)
}
