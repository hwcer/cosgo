package times

import "testing"

func TestCycle_Cycle(t *testing.T) {

	v := 2025070108
	t.Logf("===============%d================", v)
	ts, _ := ParseExpireTypeCustomize(v)
	//ts.SetTimeReset(5 * 3600)
	c1 := ts.Cycle(ExpireTypeMonthly, 3)

	s, _ := c1.Start()
	e, _ := c1.Expire()

	t.Logf("开始时间:%v", ts.String())
	_, n := c1.Cycle()
	t.Logf("当前届数:%v", n)
	t.Logf("本届开始:%v", s.String())
	t.Logf("本届结束:%v", e.String())

	v = 2025101205
	t.Logf("===============%d================", v)

	ts, _ = ParseExpireTypeCustomize(v)
	c1 = ts.Cycle(ExpireTypeDaily, 3)

	s, _ = c1.Start()
	e, _ = c1.Expire()

	t.Logf("开始时间:%v", ts.String())
	_, n = c1.Cycle()
	t.Logf("当前届数:%v", n)
	t.Logf("本届开始:%v", s.String())
	t.Logf("本届结束:%v", e.String())
}
