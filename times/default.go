package times

import (
	"time"
)

var Default *Times

func init() {
	Default = &Times{}
	Default.layout = "2006-01-02 15:04:05-0700"
	Default.WeekStartDay = 1
}

func New(v time.Time) *Times {
	return Default.New(v)
}

func Now() *Times {
	return Default.New(time.Now())
}
func Add(d time.Duration) *Times {
	return Default.Add(d)
}

func AddDate(years int, months int, days int) *Times {
	return Default.AddDate(years, months, days)
}

func Unix(v int64) *Times {
	return Default.Unix(v)
}
func Milli(v int64) *Times {
	return Default.Milli(v)
}

func Parse(value string, layout ...string) (*Times, error) {
	return Default.Parse(value, layout...)
}

func Sign(addDays int) (sign int32, str string) {
	return Default.Sign(addDays)
}

func Format(layout ...string) string {
	return Default.Format(layout...)
}

func String() string {
	return Default.String()
}

// Daily 获取一天的开始时间
// addDays：天偏移，0：今天凌晨,1:明天凌晨
// args :时,分,秒,毫秒
func Daily(addDays int) *Times {
	return Default.Daily(addDays)
}

// Weekly 获取本周开始时间
// addWeeks：周偏移，0：本周,1:下周 -1:上周
func Weekly(addWeeks int) *Times {
	return Default.Weekly(addWeeks)
}

// Monthly 获取本月开始时间
// addMonth：月偏移，0：本月,1:下月 -1:上月
func Monthly(addMonth int) *Times {
	return Default.Monthly(addMonth)
}
func Start(t ExpireType, v int) (ts *Times, err error) {
	return Default.Start(t, v)
}
func Expire(t ExpireType, v int) (ts *Times, err error) {
	return Default.Expire(t, v)
}

func SetTimeReset(v int64) {
	Default.SetTimeReset(v)
}

func GetTimeZone() string {
	return Default.GetTimeZone()
}
