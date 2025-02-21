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
	time         time.Time //时间
	layout       string    //日期输入格式，必须带时区-0700,默认精确到秒
	timeZone     string    //时区偏移量 -0700
	timeReset    [3]int    //每日重置设置，时分秒
	WeekStartDay int       //每周开始时间,默认周一   1:周一，0:周日
}

func (this *Times) New(v ...time.Time) *Times {
	t := *this
	if len(v) > 0 && !v[0].IsZero() {
		t.time = v[0]
	}
	return &t
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

func (this *Times) Unix() int64 {
	return this.Now().Unix()
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
func (this *Times) Timestamp(v int64) *Times {
	t := time.Unix(v, 0)
	return this.New(t)
}

func (this *Times) Sign(addDays int) (sign int32, str string) {
	t := this.Now()
	if addDays > 0 {
		t = t.AddDate(0, 0, addDays)
	}
	str = t.Format(signLayout)
	ret, _ := strconv.ParseUint(str, 10, 32)
	sign = int32(ret)
	return
}

func (this *Times) TimeReset(v [3]int) {
	this.timeReset = v
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
	r := time.Date(t.Year(), t.Month(), t.Day(), this.timeReset[0], this.timeReset[1], this.timeReset[2], 0, t.Location())
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

	r := time.Date(t.Year(), t.Month(), t.Day(), this.timeReset[0], this.timeReset[1], this.timeReset[2], 0, t.Location())
	if addWeeks != 0 {
		r = r.AddDate(0, 0, addWeeks*7)
	}
	return this.New(r)
}

// Monthly 获取本月开始时间
// addMonth：月偏移，0：本月,1:下月 -1:上月
func (this *Times) Monthly(addMonth int) *Times {
	t := this.Now()
	r := time.Date(t.Year(), t.Month(), 1, this.timeReset[0], this.timeReset[1], this.timeReset[2], 0, t.Location())
	if addMonth != 0 {
		r = r.AddDate(0, addMonth, 0)
	}
	return this.New(r)
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
		s = this.Daily(0).Unix()
	case ExpireTypeWeekly:
		s = this.Weekly(0).Unix()
	case ExpireTypeMonthly:
		s = this.Monthly(0).Unix()
	case ExpireTypeSecond:
		s = this.Unix()
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
		if ttl, err = this.Parse(fmt.Sprintf("%v%v", v, this.GetTimeZone()), "20060102-0700"); err == nil {
			ttl = ttl.Daily(1).Add(-1)
		}
	default:
		ttl = this.New(time.Unix(0, 0))
	}
	return
}
