package utils

import (
	"strconv"
	"time"
)

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

//获取一天的开始时间,day：天偏移，0：今天凌晨,1:明天凌晨
//args:  addDays,dayResetHour
func GetTimeDayStart(t time.Time, args ...int32) time.Time {
	var addDays, dayResetHour int

	if len(args) > 0 {
		addDays = int(args[0])
	}
	if len(args) > 1 {
		dayResetHour = int(args[1])
	}

	if dayResetHour > 0 && t.Hour() < dayResetHour {
		t = t.AddDate(0, 0, -1) //不到凌晨6点算前一天
	}

	//fmt.Printf("addDays:%v,dayResetHour:%v\n", addDays, dayResetHour)

	r := time.Date(t.Year(), t.Month(), t.Day(), dayResetHour, 0, 0, 0, t.Location())

	if addDays > 0 {
		r = r.AddDate(0, 0, addDays)
	}
	return r
}

//获取本周开始时间
//args:  addWeeks,dayResetHour
func GetTimeWeekStart(t time.Time, args ...int32) time.Time {
	var addWeeks, dayResetHour int
	if len(args) > 0 {
		addWeeks = int(args[0])
	}
	if len(args) > 1 {
		dayResetHour = int(args[1])
	}
	if dayResetHour > 0 && t.Hour() < dayResetHour {
		t = t.AddDate(0, 0, -1) //不到凌晨6点算前一天
	}

	week := int(t.Weekday())
	if week == 0 {
		week = 7
	}
	if week > 1 {
		t = t.AddDate(0, 0, -(week - 1)) //定位到周一
	}
	r := time.Date(t.Year(), t.Month(), t.Day(), dayResetHour, 0, 0, 0, t.Location())
	if addWeeks > 0 {
		r = r.AddDate(0, 0, addWeeks*7)
	}
	return r
}
