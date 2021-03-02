package utils

import (
	"strconv"
	"time"
)

//每周开始时间 1:周一，0:周日
var WeekStartDay = 1

//通过时间戳获取SIGN
func GetSignFromStamp(s int64) int32 {
	var t time.Time
	if s == 0 {
		t = time.Now()
	} else {
		t = time.Unix(s, 0)
	}
	sign, _ := strconv.Atoi(t.Format("20060102"))
	return int32(sign)
}

func GetSignFromTime(t time.Time) int32 {
	sign, _ := strconv.Atoi(t.Format("20060102"))
	return int32(sign)
}

func GetTimeFromSign(sign int32) time.Time {
	s := strconv.Itoa(int(sign))
	year, _ := strconv.Atoi(s[0:4])
	month, _ := strconv.Atoi(s[4:6])
	day, _ := strconv.Atoi(s[6:8])
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
}

//获取一天的开始时间
//addDays：天偏移，0：今天凌晨,1:明天凌晨
func GetDayTime(t time.Time, addDays int) time.Time {
	r := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	if addDays != 0 {
		r = r.AddDate(0, 0, addDays)
	}
	return r
}

//获取本周开始时间
//addWeeks：周偏移，0：本周,1:下周 -1:上周
func GetWeekTime(t time.Time, addWeeks int) time.Time {
	var addDay int
	week := int(t.Weekday())

	if week > WeekStartDay {
		addDay = -(week - WeekStartDay)
	} else if week < WeekStartDay {
		addDay = WeekStartDay - week - 7
	}
	if addDay != 0 {
		t = t.AddDate(0, 0, addDay)
	}

	r := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	if addWeeks != 0 {
		r = r.AddDate(0, 0, addWeeks*7)
	}
	return r
}
