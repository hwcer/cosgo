package utils

import (
	"fmt"
	"strconv"
	"time"
)

var Time *dateTime

const (
	DaySecond int64 = 24 * 60 * 60
)

func init() {
	Time = &dateTime{}
	//Time.time = time.Now()
	Time.zone = time.Now().Format("-0700")
	Time.reset = []int{0, 0, 0, 0}
	Time.layout = "2006-01-02 15:04:05-0700"
	Time.WeekStartDay = 1
}

type dateTime struct {
	time         time.Time //时间
	zone         string    //时区偏移量 -0700
	reset        []int     //时间重置设置，时分秒毫秒
	layout       string    //日期输入格式，必须带时区-0700,默认精确到秒
	WeekStartDay int       //每周开始时间,默认周一   1:周一，0:周日
}

func (this *dateTime) New(t time.Time) *dateTime {
	dt := *this
	dt.time = t
	dt.zone = dt.time.Format("-0700")
	return &dt
}

//Now 获取dateTime中的当前时间，默认TIME 未设置当前时间则返回系统当前时间
func (this *dateTime) Now() time.Time {
	if this.time.IsZero() {
		return time.Now()
	} else {
		return this.time
	}
}

func (this *dateTime) Unix() int64 {
	return this.time.Unix()
}

func (this *dateTime) Reset(args ...int) {
	if len(args) < 4 {
		args = append(args, 0, 0, 0, 0)
	}
	this.reset = args
}

func (this *dateTime) Layout(layout string) *dateTime {
	this.layout = layout
	return this
}

func (this *dateTime) Format() string {
	return this.Now().Format(this.layout)
}

//Daily 获取一天的开始时间
//addDays：天偏移，0：今天凌晨,1:明天凌晨
//args :时,分,秒,毫秒
func (this *dateTime) Daily(addDays int) time.Time {
	t := this.Now()
	r := time.Date(t.Year(), t.Month(), t.Day(), this.reset[0], this.reset[1], this.reset[2], this.reset[3], t.Location())
	if addDays != 0 {
		r = r.AddDate(0, 0, addDays)
	}
	return r
}

//Weekly 获取本周开始时间
//addWeeks：周偏移，0：本周,1:下周 -1:上周
func (this *dateTime) Weekly(addWeeks int) time.Time {
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

	r := time.Date(t.Year(), t.Month(), t.Day(), this.reset[0], this.reset[1], this.reset[2], this.reset[3], t.Location())
	if addWeeks != 0 {
		r = r.AddDate(0, 0, addWeeks*7)
	}
	return r
}

//Monthly 获取本月开始时间
//addMonth：月偏移，0：本月,1:下月 -1:上月
func (this *dateTime) Monthly(addMonth int) time.Time {
	t := this.Now()
	r := time.Date(t.Year(), t.Month(), 1, this.reset[0], this.reset[1], this.reset[2], this.reset[3], t.Location())
	if addMonth > 0 {
		r = r.AddDate(0, addMonth, 0)
	}
	return r
}

/*
Expire 有效期
0不限制 返回0 无刷新时间
1 日刷新  刷新礼包时间可配置具体几日
2 周刷新  刷新礼包时间可配置具体几周
3 月刷新  刷新礼包时间可配置具体几月
4 按当天0点时间   刷新礼包时间配置秒数
5 具体到期时间 20210101235959  //年月日时分秒
v = 1 :当天，周，月晚上24点
*/
func (this *dateTime) Expire(t, v int) (ttl time.Time, err error) {
	switch t {
	case 1:
		ttl = this.Daily(v)
	case 2:
		ttl = this.Weekly(v)
	case 3:
		ttl = this.Monthly(v)
	case 4:
		ttl = this.Daily(0).Add(time.Second * time.Duration(v))
	case 5:
		ttl, err = time.Parse("20060102   150405-0700", fmt.Sprintf("%v%v", v, this.zone))
	}
	return
}

func (this *dateTime) Sign(addDays int) (sign int32, str string) {
	t := this.Daily(addDays)
	format := "20060102"
	str = t.Format(format)
	ret, _ := strconv.ParseUint(str, 10, 32)
	sign = int32(ret)
	return
}
