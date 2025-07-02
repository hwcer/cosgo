package times

import (
	"github.com/hwcer/cosgo/utils"
	"time"
)

type Cycle struct {
	*Times
	t ExpireType
	v int
}

func NewCycle(ts *Times, t ExpireType, v int) *Cycle {
	if v == 0 {
		v = 1
	}
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
	if this.v == 1 || !this.Maybe() {
		return Now().Start(this.t, this.v)
	}
	switch this.t {
	case ExpireTypeDaily:
		r = this.Daily(0)
		n := this.secondCycle(r, 86400*this.v)
		r = r.AddDate(0, 0, (n-1)*this.v)
	case ExpireTypeWeekly:
		r = this.Weekly(0)
		n := this.secondCycle(r, 86400*this.v*7)
		r = r.AddDate(0, 0, (n-1)*this.v*7)
	case ExpireTypeMonthly:
		r = this.Monthly(0)
		n := this.monthlyCycle(r)
		r = r.AddDate(0, (n-1)*this.v, 0)
	case ExpireTypeSecond:
		r = this.Times
		n := this.secondCycle(r, this.v)
		r = r.Add(time.Duration((n-1)*this.v) * time.Second)
	}
	return
}

// Expire 本届结束时间
func (this *Cycle) Expire() (r *Times, err error) {
	if this.v == 1 || !this.Maybe() {
		return Now().Expire(this.t, this.v)
	}
	switch this.t {
	case ExpireTypeDaily:
		r = this.Daily(0)
		n := this.secondCycle(r, 86400*this.v)
		r = r.AddDate(0, 0, n*this.v)
	case ExpireTypeWeekly:
		r = this.Weekly(0)
		n := this.secondCycle(r, 86400*this.v*7)
		r = r.AddDate(0, 0, n*this.v*7)
	case ExpireTypeMonthly:
		r = this.Monthly(0)
		n := this.monthlyCycle(r)
		r = r.AddDate(0, n*this.v, 0)
	case ExpireTypeSecond:
		r = this.Times
		n := this.secondCycle(r, this.v)
		r = r.Add(time.Duration(n*this.v) * time.Second)
	}
	if r != nil {
		r = r.Add(-1)
	}
	
	return
}

// Cycle 当前是第几届，1开始
func (this *Cycle) Cycle() (era *Times, r int) {
	switch this.t {
	case ExpireTypeDaily:
		era = this.Daily(0)
		r = this.secondCycle(era, 86400*this.v)
	case ExpireTypeWeekly:
		era = this.Weekly(0)
		r = this.secondCycle(era, 86400*this.v*7)
	case ExpireTypeMonthly:
		era = this.Monthly(0)
		r = this.monthlyCycle(era)
	case ExpireTypeSecond:
		era = this.Times
		r = this.secondCycle(era, this.v)
	}
	return
}

func (this *Cycle) secondCycle(era *Times, n int) (r int) {
	s := era.Now().Unix()
	t := time.Now().Unix()
	return utils.Ceil(int(t-s), n)
}

// monthlyCycle 计算到当前时间累计有几个月，包含开始时间点，和当前月
func (this *Cycle) monthlyCycle(era *Times) int {
	s := era.Now()
	t := time.Now()
	m1, m2 := int(s.Month()), int(t.Month())
	var r int
	if y1, y2 := s.Year(), t.Year(); y1 < y2 {
		r += 12 - m1 + 1
		for i := y1 + 1; i < y2; i++ {
			r += 12
		}
	}
	r += m2
	return utils.Ceil(r, this.v)
}
