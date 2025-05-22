package times

import (
	"time"
)

var times *Times

func init() {
	times = &Times{}
	times.layout = "2006-01-02 15:04:05-0700"
	times.WeekStartDay = 1
}

func New(v time.Time) *Times {
	return times.New(v)
}

func Now() time.Time {
	return times.Now()
}
func Add(d time.Duration) *Times {
	return times.Add(d)
}

func AddDate(years int, months int, days int) *Times {
	return times.AddDate(years, months, days)
}

func Unix(v int64) *Times {
	return times.Unix(v)
}
func Milli(v int64) *Times {
	return times.Milli(v)
}

func Parse(value string, layout ...string) (*Times, error) {
	return times.Parse(value, layout...)
}

func Sign(addDays int) (sign int32, str string) {
	return times.Sign(addDays)
}

func Format(layout ...string) string {
	return times.Format(layout...)
}

func String() string {
	return times.String()
}

// Daily 获取一天的开始时间
// addDays：天偏移，0：今天凌晨,1:明天凌晨
// args :时,分,秒,毫秒
func Daily(addDays int) *Times {
	return times.Daily(addDays)
}

// Weekly 获取本周开始时间
// addWeeks：周偏移，0：本周,1:下周 -1:上周
func Weekly(addWeeks int) *Times {
	return times.Weekly(addWeeks)
}

// Monthly 获取本月开始时间
// addMonth：月偏移，0：本月,1:下月 -1:上月
func Monthly(addMonth int) *Times {
	return times.Monthly(addMonth)
}

// Verify 验证是否有效,true:有效
func Verify(et ExpireType, v int64) (r bool) {
	return times.Verify(et, v)
}

func Expire(t ExpireType, v int) (ttl *Times, err error) {
	return times.Expire(t, v)
}

func SetTimeReset(v int64) {
	times.SetTimeReset(v)
}

func SetTimeZone(zone string) {
	times.SetTimeZone(zone)
}
func GetTimeZone() string {
	return times.GetTimeZone()
}
