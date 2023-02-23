package utils

import (
	"fmt"
	"strconv"
	"time"
)

var Time *DateTime

const (
	DaySecond int64 = 24 * 60 * 60
)

func init() {
	Time = &DateTime{}
	//Time.time = time.Now()
	Time.timeZone = time.Now().Format("-0700")
	Time.timeReset = []int{0, 0, 0, 0}
	Time.layout = "2006-01-02 15:04:05-0700"
	Time.WeekStartDay = 1
}

type DateTime struct {
	time         time.Time //时间
	layout       string    //日期输入格式，必须带时区-0700,默认精确到秒
	timeZone     string    //时区偏移量 -0700
	timeReset    []int     //时间重置设置，时分秒毫秒
	WeekStartDay int       //每周开始时间,默认周一   1:周一，0:周日
}

func (this *DateTime) New(t time.Time) *DateTime {
	dt := *this
	dt.time = t
	dt.timeZone = dt.time.Format("-0700")
	return &dt
}

func (this *DateTime) Now() time.Time {
	if this.time.IsZero() {
		return time.Now()
	} else {
		return this.time
	}
}
func (this *DateTime) Add(d time.Duration) *DateTime {
	t := this.Now()
	t = t.Add(d)
	return this.New(t)
}

func (this *DateTime) AddDate(years int, months int, days int) *DateTime {
	t := this.Now()
	t = t.AddDate(years, months, days)
	return this.New(t)
}

func (this *DateTime) Unix() int64 {
	return this.Now().Unix()
}

func (this *DateTime) Parse(value string, layout ...string) (*DateTime, error) {
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

func (this *DateTime) TimeZone(zone string) {
	this.timeZone = zone
}

func (this *DateTime) TimeReset(args ...int) {
	if len(args) < 4 {
		args = append(args, 0, 0, 0, 0)
	}
	this.timeReset = args
}

//func (this *DateTime) Layout(layout string) *DateTime {
//	this.layout = layout
//	return this
//}

func (this *DateTime) Format(layout ...string) string {
	format := this.layout
	if len(layout) > 0 {
		format = layout[0]
	}
	return this.Now().Format(format)
}

// Daily 获取一天的开始时间
// addDays：天偏移，0：今天凌晨,1:明天凌晨
// args :时,分,秒,毫秒
func (this *DateTime) Daily(addDays int) *DateTime {
	t := this.Now()
	r := time.Date(t.Year(), t.Month(), t.Day(), this.timeReset[0], this.timeReset[1], this.timeReset[2], this.timeReset[3], t.Location())
	if addDays != 0 {
		r = r.AddDate(0, 0, addDays)
	}
	return this.New(r)
}

// Weekly 获取本周开始时间
// addWeeks：周偏移，0：本周,1:下周 -1:上周
func (this *DateTime) Weekly(addWeeks int) *DateTime {
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

	r := time.Date(t.Year(), t.Month(), t.Day(), this.timeReset[0], this.timeReset[1], this.timeReset[2], this.timeReset[3], t.Location())
	if addWeeks != 0 {
		r = r.AddDate(0, 0, addWeeks*7)
	}
	return this.New(r)
}

// Monthly 获取本月开始时间
// addMonth：月偏移，0：本月,1:下月 -1:上月
func (this *DateTime) Monthly(addMonth int) *DateTime {
	t := this.Now()
	r := time.Date(t.Year(), t.Month(), 1, this.timeReset[0], this.timeReset[1], this.timeReset[2], this.timeReset[3], t.Location())
	if addMonth > 0 {
		r = r.AddDate(0, addMonth, 0)
	}
	return this.New(r)
}

/*
Expire 有效期
0不限制 返回0 无刷新时间
1 日刷新  刷新礼包时间可配置具体几日
2 周刷新  刷新礼包时间可配置具体几周
3 月刷新  刷新礼包时间可配置具体几月
4 按当天0点时间   刷新礼包时间配置秒数
5 具体到期时间 2021010123  //年月日时,精确到小时
v = 1 :当天，周，月晚上24点
*/

const (
	DateTimeExpireNone      int = 0
	DateTimeExpireDaily         = 1
	DateTimeExpireWeekly        = 2
	DateTimeExpireMonthly       = 3
	DateTimeExpireSecond        = 4
	DateTimeExpireCustomize     = 5
)

func (this *DateTime) Expire(t, v int) (ttl *DateTime, err error) {
	switch t {
	case DateTimeExpireDaily:
		ttl = this.Daily(v)
	case DateTimeExpireWeekly:
		ttl = this.Weekly(v)
	case DateTimeExpireMonthly:
		ttl = this.Monthly(v)
	case DateTimeExpireSecond:
		ttl = this.Daily(0).Add(time.Second * time.Duration(v))
	case DateTimeExpireCustomize:
		ttl, err = this.Parse(fmt.Sprintf("%v%v", v, this.timeZone), "2006010215-0700")
	default:
		ttl = this.New(time.Unix(0, 0))
	}
	return
}

func (this *DateTime) Sign(addDays int) (sign int32, str string) {
	t := this.Daily(addDays)
	format := "20060102"
	str = t.Format(format)
	ret, _ := strconv.ParseUint(str, 10, 32)
	sign = int32(ret)
	return
}

// STime 开始时间 0,1：今天凌晨,2:明天凌晨(第二天)
func (this *DateTime) STime(v int) *DateTime {
	if v <= 0 {
		v = 0
	} else {
		v -= 1
	}
	return this.Daily(v)
}

// ETime 结束时间 0,1：今天24点,2:明天24点(第二天结束时间)
func (this *DateTime) ETime(v int) *DateTime {
	t := this.STime(v)
	t = t.AddDate(0, 0, 1)
	return t
}

func (this *DateTime) Timestamp(v int64) *DateTime {
	t := time.Unix(v, 0)
	return this.New(t)
}
