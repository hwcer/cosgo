package times

import (
	"testing"
	"time"
)

func TestGetDayStartTime(t *testing.T) {
	SetTimeReset(6 * 3600) //设置每日6点切换到下一天，不改变时区
	t.Logf("当前时区:%v", GetTimeZone())
	t.Logf("%v", Daily(0).String())
	t.Logf("%v", Daily(1).String())
	t.Logf("Weekly:%v", Weekly(0).String())
	t.Logf("Monthly:%v", Monthly(0).String())
	v := 20231229
	if expire, err := Expire(5, v); err != nil {
		t.Logf("ERR:%v", err)
	} else {
		t.Logf("Expire:%v", expire.String())
	}

	t.Logf("NOW:%v", String())

	x := Daily(0)
	t.Logf("Daily: %v,  %v", x.Now().Unix(), x.String())
	if expire, err := x.Expire(ExpireTypeDaily, 1); err != nil {
		t.Logf("ERR:%v", err)
	} else {
		t.Logf("Expire:%v  %v", expire.Now().Unix(), expire.String())
	}

	y := x.Now().Unix() + 86400
	z := x.New(time.Unix(y, 0))
	sign, _ := z.Sign(0)
	t.Logf("Daily:%v  sign:%v", z.String(), sign)

	if s, err := Parse("2023-02-28 06:00:00+0800"); err != nil {
		t.Logf("ERR:%v", err)
	} else {
		t.Logf("Parse:%v", s.String())
		sign, _ = s.Sign(0)
		t.Logf("sign:%v", sign)
		t.Logf("Daily:%v", s.Daily(0).String())
		ttl, _ := s.Expire(ExpireTypeDaily, 1)
		t.Logf("Expire:%v", ttl.String())
	}

}
