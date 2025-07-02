package times

import "testing"

func TestCycle_Cycle(t *testing.T) {
	v := 2025010315
	ts, _ := ParseExpireTypeCustomize(v)
	c1 := ts.Cycle(ExpireTypeMonthly, 2)

	s, _ := c1.Start()
	e, _ := c1.Expire()

	t.Logf("开始时间:%v", ts.String())
	_, n := c1.Cycle()
	t.Logf("当前届数:%v", n)
	t.Logf("本届开始:%v", s.String())
	t.Logf("本届结束:%v", e.String())

	v = 2025070115
	ts, _ = ParseExpireTypeCustomize(v)
	c1 = ts.Cycle(ExpireTypeDaily, 2)

	s, _ = c1.Start()
	e, _ = c1.Expire()

	t.Logf("开始时间:%v", ts.String())
	_, n = c1.Cycle()
	t.Logf("当前届数:%v", n)
	t.Logf("本届开始:%v", s.String())
	t.Logf("本届结束:%v", e.String())
}
