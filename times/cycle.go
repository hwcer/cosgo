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

// monthlyCycle 计算到当前时间累计有几个月，包含开始时间点，和当前月
// 0开始,第一届为0
func (this *Cycle) monthlyCycle(era *Times) int {
	// 计算总月数差
	y1, y2 := era.Now().Year(), time.Now().Year()
	m1, m2 := int(era.Now().Month()), int(time.Now().Month())
	r := (y2-y1)*12 + (m2 - m1)

	// 考虑日期因素：如果当前日期早于开始日期，月数减1
	if time.Now().Day() < era.Now().Day() {
		r--
	}

	// 计算届数，从0开始计数
	v := r / this.v
	return v
}
