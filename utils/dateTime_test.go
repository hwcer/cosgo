package utils

import (
	"testing"
)

var timeFormat = "2006-01-02 15:04:05.999999999 -0700"

func TestGetDayStartTime(t *testing.T) {

	t.Logf("%v", Time.Daily(0).Format(timeFormat))
	t.Logf("%v", Time.Daily(1).Format(timeFormat))
	v := 20211229115830
	e, _ := Time.Expire(5, v)

	t.Logf("Expire:%v", e.String())
}
