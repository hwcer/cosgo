package times

import (
	"testing"
	"time"
)

func TestGetDayStartTime(t *testing.T) {

	t.Logf("%v", Daily(0).String())
	t.Logf("%v", Daily(1).String())
	v := 2023122911
	if expire, err := Expire(5, v); err != nil {
		t.Logf("ERR:%v", err)
	} else {
		t.Logf("Expire:%v", expire.String())
	}

	t.Logf("NOW:%v", String())

	x := Daily(0)
	t.Logf("Daily: %v,  %v", x.Unix(), x.String())
	if expire, err := x.Expire(ExpireTypeDaily, 1); err != nil {
		t.Logf("ERR:%v", err)
	} else {
		t.Logf("Expire:%v  %v", expire.Unix(), expire.String())
	}

	y := x.Unix() + 86400
	z := x.New(time.Unix(y, 0))
	sign, _ := z.Sign(0)
	t.Logf("Daily:%v  sign:%v", z.String(), sign)

	if s, err := Parse("2023-02-28 24:00:00+0800"); err != nil {
		t.Logf("ERR:%v", err)
	} else {
		t.Logf("Parse:%v", s.String())
	}

}
