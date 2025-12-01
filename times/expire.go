package times

/*
Expire 有效期
0不限制 返回0 无刷新时间
1 日刷新    可配置具体几日
2 周刷新    可配置具体几周
3 月刷新    可配置具体几月
4 当前时间之后Y秒开始
5 具体到期日期 2006010215  //年月日时,精确到当前23:59:59
6 直接使用时间戳
v = 1 :当天，周，月 23:59:59
*/

type ExpireType int8

const (
	ExpireTypeNone      ExpireType = 0
	ExpireTypeDaily     ExpireType = 1
	ExpireTypeWeekly    ExpireType = 2
	ExpireTypeMonthly   ExpireType = 3
	ExpireTypeSecond    ExpireType = 4
	ExpireTypeCustomize ExpireType = 5
	ExpireTimeTimeStamp ExpireType = 6
)

func (t ExpireType) Has() bool {
	return t >= ExpireTypeNone && t <= ExpireTimeTimeStamp
}
