package values

type Paging struct {
	Page int      //当前页
	Size  int      //每页大小
	Total int     //总记录数
	Rows interface{}
}