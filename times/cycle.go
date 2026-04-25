package times

import (
	"time"
)

type Cycle struct {
	*Times
	t ExpireType
	v int
}

func NewCycle(ts *Times, t ExpireType, v int) *Cycle {
	return &Cycle{Times: ts, t: t, v: v}
}

// Era 活动开始时间
func (this *Cycle) Era() *Times {
	switch this.t {
	case ExpireTypeDaily:
		return this.Times.Daily(0)
	case ExpireTypeWeekly:
		return this.Times.Weekly(0)
	case ExpireTypeMonthly:
		return this.Times.Monthly(0)
	case ExpireTypeSecond:
		return this.Times
	default:
		return this.Times
	}
}

func (this *Cycle) Maybe() bool {
	return this.t == ExpireTypeDaily || this.t == ExpireTypeWeekly || this.t == ExpireTypeMonthly || this.t == ExpireTypeSecond
}

// Start 当前届开始时间
func (this *Cycle) Start() (r *Times, err error) {
	if this.v <= 1 || !this.Maybe() {
		return Default.Start(this.t, this.v)
	}
	switch this.t {
	case ExpireTypeDaily:
		r = this.Daily(0)
		n := this.secondCycle(r, 86400*this.v)
		if n != 0 {
			r = r.AddDate(0, 0, n*this.v)
		}
	case ExpireTypeWeekly:
		r = this.Weekly(0)
		n := this.secondCycle(r, 86400*this.v*7)
		if n != 0 {
			r = r.AddDate(0, 0, n*this.v*7)
		}
	case ExpireTypeMonthly:
		r = this.Monthly(0)
		n := this.monthlyCycle(r)
		if n != 0 {
			r = r.AddDate(0, n*this.v, 0)
		}

	case ExpireTypeSecond:
		r = this.Times
		n := this.secondCycle(r, this.v)
		if n != 0 {
			r = r.Add(time.Duration(n*this.v) * time.Second)
		}

	}
	return
}

// Expire 本届结束时间
func (this *Cycle) Expire() (r *Times, err error) {
	if this.v <= 1 || !this.Maybe() {
		return Default.Expire(this.t, this.v)
	}
	switch this.t {
	case ExpireTypeDaily:
		r = this.Daily(0)
		n := this.secondCycle(r, 86400*this.v)
		if diff := n + 1; diff != 0 {
			r = r.AddDate(0, 0, diff*this.v)
		}

	case ExpireTypeWeekly:
		r = this.Weekly(0)
		n := this.secondCycle(r, 86400*this.v*7)
		if diff := n + 1; diff != 0 {
			r = r.AddDate(0, 0, diff*this.v*7)
		}

	case ExpireTypeMonthly:
		r = this.Monthly(0)
		n := this.monthlyCycle(r)
		if diff := n + 1; diff != 0 {
			r = r.AddDate(0, diff*this.v, 0)
		}
	case ExpireTypeSecond:
		r = this.Times
		n := this.secondCycle(r, this.v)
		if diff := n + 1; diff != 0 {
			r = r.Add(time.Duration(diff*this.v) * time.Second)
		}
	}
	if r != nil {
		r = r.Add(-1) //减去1纳秒确保时间点回到本届最后一纳秒，而不是下一届的第一纳秒
	}
	return
}

// Cycle 当前是第几届，0开始
func (this *Cycle) Cycle() (era *Times, r int) {
	switch this.t {
	case ExpireTypeDaily:
		era = this.Daily(0)
		r = this.secondCycle(era, 86400*this.v)
	case ExpireTypeWeekly:
		era = this.Weekly(0)
		r = this.secondCycle(era, int(WeekSecond)*this.v)
	case ExpireTypeMonthly:
		era = this.Monthly(0)
		r = this.monthlyCycle(era)
	case ExpireTypeSecond:
		era = this.Times
		r = this.secondCycle(era, this.v)
	}
	return
}

// secondCycle 按秒计算当前是第几轮
// 0开始,第一届为0
func (this *Cycle) secondCycle(era *Times, n int) (r int) {
	s := era.Now().Unix()
	t := time.Now().Unix()
	return int(t-s) / n
}

// monthlyCycle 计算从 era 到现在累计的月数差,从 0 开始计届。
//
// 时区说明:
//   - `era.Now()` 和 `time.Now()` 都使用 time.Local 解释 Year/Month/Day。
//   - 若 era 记录时与查询时进程所在时区不同(例如生产/测试服位于不同时区),
//     月份/日期切换边界可能偏移 1 天,进而偏移 1 个月的届数。
//   - DST(夏令时)切换影响在小时级,不影响 Month/Day 计算。
//   - 如果调用方需要跨时区确定性,应保证 era 和查询方同处 time.Local,
//     或 fork 本函数改为显式在 UTC 下计算。
func (this *Cycle) monthlyCycle(era *Times) int {
	eraTime := era.Now()
	now := time.Now()
	// 用单次 time.Now() 快照，避免 Year/Month/Day 分别调用时跨秒/跨日
	r := (now.Year()-eraTime.Year())*12 + int(now.Month()) - int(eraTime.Month())
	if now.Day() < eraTime.Day() {
		r--
	}
	return r / this.v
}
