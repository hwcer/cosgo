package xlsx

var Config = &struct {
	Suffix  string //表名结尾
	Package string //包名
	Summary string //总表名,留空不生成总表
}{
	Suffix:  "Row",
	Package: "Pb",
	Summary: "StaticData",
}
