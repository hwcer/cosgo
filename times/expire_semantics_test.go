package times

import (
	"testing"
	"time"
)

// TestExpireIsIntervalEnd 钉住 Expire 的左闭右开语义:
// 返回的是区间【右端点】本身,该时间点不属于本届,判定有效一律 now < expire。
//
// 历史上此处返回 Daily(v).Add(-1)(本届最后一纳秒),调用方取 .Unix() 秒截断后
// 会得到 23:59:59,令 now < expire 的判定提前 1 秒过期。
func TestExpireIsIntervalEnd(t *testing.T) {
	SetTimeReset(0) //零点日切,避免受全局配置影响
	base, err := Parse("2026-08-19 10:30:00+0800")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	cases := []struct {
		name string
		et   ExpireType
		want *Times
	}{
		{"daily", ExpireTypeDaily, base.Daily(1)},
		{"weekly", ExpireTypeWeekly, base.Weekly(1)},
		{"monthly", ExpireTypeMonthly, base.Monthly(1)},
	}
	for _, c := range cases {
		got, e := base.Expire(c.et, 1)
		if e != nil {
			t.Fatalf("%s: Expire err %v", c.name, e)
		}
		if !got.Now().Equal(c.want.Now()) {
			t.Errorf("%s: Expire=%v, 应等于区间右端点 %v", c.name, got.String(), c.want.String())
		}
		//右端点本身不属于本届:起点严格早于终点,且终点整秒不再算作有效
		if !base.Now().Before(got.Now()) {
			t.Errorf("%s: 起点 %v 应严格早于终点 %v", c.name, base.String(), got.String())
		}
	}
}

// TestExpireDailyIsNextDayZero 每日过期点应恰为次日 0 点,而非当日 23:59:59
func TestExpireDailyIsNextDayZero(t *testing.T) {
	SetTimeReset(0)
	base, err := Parse("2026-08-19 10:30:00+0800")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got, err := base.Expire(ExpireTypeDaily, 1)
	if err != nil {
		t.Fatalf("Expire: %v", err)
	}
	n := got.Now()
	if n.Hour() != 0 || n.Minute() != 0 || n.Second() != 0 || n.Nanosecond() != 0 {
		t.Errorf("Expire(Daily,1)=%v, 应为次日 00:00:00 整", got.String())
	}
	if d := n.Sub(base.Daily(0).Now()); d != 24*time.Hour {
		t.Errorf("本届时长 %v, 应为 24h(区间右端点减左端点)", d)
	}
}

// TestCycleExpireIsIntervalEnd Cycle.Expire 与 Times.Expire 同口径
func TestCycleExpireIsIntervalEnd(t *testing.T) {
	SetTimeReset(0)
	base, err := Parse("2026-08-19 10:30:00+0800")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	//v=2 走 Cycle 自有分支(v<=1 会退回 Default.Expire)
	got, err := base.Cycle(ExpireTypeDaily, 2).Expire()
	if err != nil {
		t.Fatalf("Cycle.Expire: %v", err)
	}
	n := got.Now()
	if n.Hour() != 0 || n.Minute() != 0 || n.Second() != 0 || n.Nanosecond() != 0 {
		t.Errorf("Cycle.Expire=%v, 应为整点(区间右端点),不应是本届最后一纳秒", got.String())
	}
}
