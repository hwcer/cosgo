package times

import (
	"fmt"
	"strconv"
	"time"
)

const (
	signLayout       = "20060102"
	DateLayout       = "2006-01-02"
	DaySecond  int64 = 24 * 60 * 60
	WeekSecond int64 = DaySecond * 7
)

type Times struct {
	time         time.Time     //时间
	layout       string        //日期输入格式，必须带时区-0700,默认精确到秒
	timeZone     string        //时区偏移量 -0700
	timeReset    time.Duration //每日几点重置日(秒)
	WeekStartDay int           //每周开始时间,默认周一   1:周一，0:周日
}

func (this *Times) New(v time.Time) *Times {
	t := *this
	t.time = v
	return &t
}

func (this *Times) Unix(v int64) *Times {
	t := time.Unix(v, 0)
	return this.New(t)
}
func (this *Times) Milli(v int64) *Times {
	t := time.UnixMilli(v)
	return this.New(t)
}

func (this *Times) Now() time.Time {
	if this.time.IsZero() {
		return time.Now()
	} else {
		return this.time
	}
}
func (this *Times) Add(d time.Duration) *Times {
	t := this.Now()
	t = t.Add(d)
	return this.New(t)
}

func (this *Times) AddDate(years int, months int, days int) *Times {
	t := this.Now()
	t = t.AddDate(years, months, days)
	return this.New(t)
}

func (this *Times) Parse(value string, layout ...string) (*Times, error) {
	var lay string
	if len(layout) > 0 {
		lay = layout[0]
	} else {
		lay = this.layout
	}
	t, err := time.Parse(lay, value)
	if err != nil {
		return nil, err
	}
	return this.New(t), nil
}

func (this *Times) Sign(addDays int) (sign int32, str string) {
	t := this.Now()
	if this.timeReset != 0 {
		t = t.Add(-this.timeReset)
	}
	if addDays > 0 {
		t = t.AddDate(0, 0, addDays)
	}
	str = t.Format(signLayout)
	ret, _ := strconv.ParseUint(str, 10, 32)
	sign = int32(ret)
	return
}

func (this *Times) Format(layout ...string) string {
	format := this.layout
	if len(layout) > 0 {
		format = layout[0]
	}
	return this.Now().Format(format)
}

func (this *Times) String() string {
	return this.Format()
}

// Daily 获取一天的开始时间
// addDays：天偏移，0：今天凌晨,1:明天凌晨
// args :时,分,秒,毫秒
func (this *Times) Daily(addDays int) *Times {
	t := this.Now()
	if this.timeReset != 0 {
		t = t.Add(-this.timeReset)
	}
	r := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	r = r.Add(this.timeReset)
	if addDays != 0 {
		r = r.AddDate(0, 0, addDays)
	}
	return this.New(r)
}

// Weekly 获取本周开始时间
// addWeeks：周偏移，0：本周,1:下周 -1:上周
func (this *Times) Weekly(addWeeks int) *Times {
	var addDay int
	t := this.Now()
	week := int(t.Weekday())
	if week > this.WeekStartDay {
		addDay = -(week - this.WeekStartDay)
	} else if week < this.WeekStartDay {
		addDay = this.WeekStartDay - week - 7
	}
	if addDay != 0 {
		t = t.AddDate(0, 0, addDay)
	}
	r := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	if this.timeReset != 0 {
		r = r.Add(this.timeReset)
	}
	if addWeeks != 0 {
		r = r.AddDate(0, 0, addWeeks*7)
	}
	return this.New(r)
}

// Monthly 获取本月开始时间
// addMonth：月偏移，0：本月,1:下月 -1:上月
func (this *Times) Monthly(addMonth int) *Times {
	t := this.Now()
	r := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	if this.timeReset != 0 {
		r = r.Add(this.timeReset)
	}
	if addMonth != 0 {
		r = r.AddDate(0, addMonth, 0)
	}
	return this.New(r)
}

// SetTimeReset 每日凌晨偏移时间，6点重置 v=6*3600
func (this *Times) SetTimeReset(v int64) {
	this.timeReset = time.Duration(v) * time.Second
}

func (this *Times) SetTimeZone(zone string) {
	this.timeZone = zone
}
func (this *Times) GetTimeZone() string {
	if this.timeZone == "" {
		this.timeZone = this.time.Format("-0700")
	}
	return this.timeZone
}

// Verify 验证是否有效,true:有效
func (this *Times) Verify(t ExpireType, v int64) (r bool) {
	if t == ExpireTypeNone {
		return true
	}
	var s int64
	switch t {
	case ExpireTypeDaily:
		s = this.Daily(0).Now().Unix()
	case ExpireTypeWeekly:
		s = this.Weekly(0).Now().Unix()
	case ExpireTypeMonthly:
		s = this.Monthly(0).Now().Unix()
	case ExpireTypeSecond:
		s = this.Now().Unix()
	default:
		return false
	}
	return v > s
}

// Expire 过期时间
func (this *Times) Expire(t ExpireType, v int) (ttl *Times, err error) {
	switch t {
	case ExpireTypeDaily:
		ttl = this.Daily(v).Add(-1)
	case ExpireTypeWeekly:
		ttl = this.Weekly(v).Add(-1)
	case ExpireTypeMonthly:
		ttl = this.Monthly(v).Add(-1)
	case ExpireTypeSecond:
		ttl = this.Add(time.Second * time.Duration(v))
	case ExpireTypeCustomize:
		s := fmt.Sprintf("%v%v", v, this.GetTimeZone())
		if ttl, err = this.Parse(s, "20060102-0700"); err == nil {
			ttl = ttl.Daily(1)
		}
	default:
		ttl = this.New(time.Unix(0, 0))
	}
	return
}
