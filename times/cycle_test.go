package times

import (
	"testing"
)

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

// TestCycle_AllTypes 测试Cycle函数在四种不同ExpireType情况下的行为
func TestCycle_AllTypes(t *testing.T) {
	// 使用当前时间作为基准
	v := 2025070108
	ts, _ := ParseExpireTypeCustomize(v)

	// 测试1: ExpireTypeDaily
	t.Logf("\n测试 ExpireTypeDaily:")
	cycleDaily := ts.Cycle(ExpireTypeDaily, 1) // 周期为1天
	eraDaily, nDaily := cycleDaily.Cycle()
	t.Logf("每日周期 - 开始时间: %v, 当前届数: %v", eraDaily.String(), nDaily)
	sDaily, _ := cycleDaily.Start()
	eDaily, _ := cycleDaily.Expire()
	t.Logf("本届开始: %v, 本届结束: %v", sDaily.String(), eDaily.String())

	// 测试2: ExpireTypeWeekly
	t.Logf("\n测试 ExpireTypeWeekly:")
	cycleWeekly := ts.Cycle(ExpireTypeWeekly, 1) // 周期为1周
	eraWeekly, nWeekly := cycleWeekly.Cycle()
	t.Logf("每周周期 - 开始时间: %v, 当前届数: %v", eraWeekly.String(), nWeekly)
	sWeekly, _ := cycleWeekly.Start()
	eWeekly, _ := cycleWeekly.Expire()
	t.Logf("本届开始: %v, 本届结束: %v", sWeekly.String(), eWeekly.String())

	// 测试3: ExpireTypeMonthly
	t.Logf("\n测试 ExpireTypeMonthly:")
	cycleMonthly := ts.Cycle(ExpireTypeMonthly, 1) // 周期为1个月
	eraMonthly, nMonthly := cycleMonthly.Cycle()
	t.Logf("每月周期 - 开始时间: %v, 当前届数: %v", eraMonthly.String(), nMonthly)
	sMonthly, _ := cycleMonthly.Start()
	eMonthly, _ := cycleMonthly.Expire()
	t.Logf("本届开始: %v, 本届结束: %v", sMonthly.String(), eMonthly.String())

	// 测试4: ExpireTypeSecond
	t.Logf("\n测试 ExpireTypeSecond:")
	cycleSecond := ts.Cycle(ExpireTypeSecond, 60) // 周期为60秒
	eraSecond, nSecond := cycleSecond.Cycle()
	t.Logf("秒周期(60秒) - 开始时间: %v, 当前届数: %v", eraSecond.String(), nSecond)
	sSecond, _ := cycleSecond.Start()
	eSecond, _ := cycleSecond.Expire()
	t.Logf("本届开始: %v, 本届结束: %v", sSecond.String(), eSecond.String())
}
