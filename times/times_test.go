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
	v := 2023122914
	if expire, err := Expire(5, v); err != nil {
		t.Logf("ERR:%v", err)
	} else {
		t.Logf("Expire:%v", expire.String())
	}

	t.Logf("NOW:%v", String())

	x := Daily(-1)
	t.Logf("Daily: %v,  %v", x.Now().Unix(), x.String())
	if expire, err := x.Expire(ExpireTypeDaily, 2); err != nil {
		t.Logf("ERR:%v", err)
	} else {
		t.Logf("Expire:%v  %v", expire.Now().Unix(), expire.String())
	}

	y := x.Now().Unix() + 86400
	z := x.New(time.Unix(y, 0))
	sign, _ := z.Sign(0)
	t.Logf("Daily:%v  sign:%v", z.String(), sign)

	if s, err := Parse("2023-02-28 05:00:00+0800"); err != nil {
		t.Logf("ERR:%v", err)
	} else {
		t.Logf("Parse:%v", s.String())
		sign, _ = s.Sign(0)
		t.Logf("sign:%v", sign)
		t.Logf("Daily:%v", s.Daily(0).String())
		ttl, _ := s.Expire(ExpireTypeDaily, 1)
		t.Logf("Expire:%v", ttl.String())
	}

	daily := Daily(0)
	sign, _ = daily.Sign(0)
	t.Logf("Daily:%v  sign:%v", daily.String(), sign)
	st, _ := daily.ParseSign(sign)
	t.Logf("ParseSign:%v ", st.String())

}

// TestSetTimeOffset 时间偏差语义:
//   - 系统时间路径(Now):系统时钟 + offset
//   - 缓存时间(New/Unix/Milli/Parse 写入)视为已偏移,不再叠加
//   - 派生 Times(New/Daily/Weekly...)继承 offset
//   - Cycle 届数按偏差后的"现在"计算
//   - 入参多态:Duration=相对偏差,int64=Unix 秒时间戳,time.Time=设成指定时刻
func TestSetTimeOffset(t *testing.T) {
	SetTimeReset(0) //零点日切,避免受全局配置影响
	t.Cleanup(func() { SetTimeOffset(time.Duration(0)) })

	//零偏差:Now 即系统时钟
	if err := SetTimeOffset(time.Duration(0)); err != nil {
		t.Fatalf("SetTimeOffset(0): %v", err)
	}
	if d := Now().Sub(time.Now()); d > 2*time.Second || d < -2*time.Second {
		t.Fatalf("零偏差 Now=%v, 与系统时钟差 %v", Now(), d)
	}

	//+1h:Duration 相对偏差,系统路径整体前拨
	if err := SetTimeOffset(time.Hour); err != nil {
		t.Fatalf("SetTimeOffset(+1h): %v", err)
	}
	want := time.Now().Add(time.Hour)
	if d := Now().Sub(want); d > 2*time.Second || d < -2*time.Second {
		t.Fatalf("+1h 偏差 Now=%v, 期望≈%v, 差 %v", Now(), want, d)
	}

	//int64 时间戳:把"现在"视为该时刻
	target := time.Now().Add(2 * time.Hour)
	if err := SetTimeOffset(target.Unix()); err != nil {
		t.Fatalf("SetTimeOffset(int64): %v", err)
	}
	if d := Now().Sub(target); d > 2*time.Second || d < -2*time.Second {
		t.Fatalf("int64 时间戳偏差 Now=%v, 期望≈%v, 差 %v", Now(), target, d)
	}

	//time.Time:设成指定时刻,自动计算偏差(回拨一天)
	target = time.Now().Add(-24 * time.Hour)
	if err := SetTimeOffset(target); err != nil {
		t.Fatalf("SetTimeOffset(time.Time): %v", err)
	}
	if d := Now().Sub(target); d > 2*time.Second || d < -2*time.Second {
		t.Fatalf("time.Time 偏差 Now=%v, 期望≈%v, 差 %v", Now(), target, d)
	}

	//非法类型:报错且偏差不变
	if err := SetTimeOffset("bad"); err == nil {
		t.Fatalf("非法入参应报错")
	}
	if d := Now().Sub(target); d > 2*time.Second || d < -2*time.Second {
		t.Fatalf("报错后偏差不应变: Now=%v, 期望≈%v, 差 %v", Now(), target, d)
	}

	//回到 +1h 继续断言缓存/派生/-cycle 语义
	if err := SetTimeOffset(time.Hour); err != nil {
		t.Fatalf("SetTimeOffset(+1h): %v", err)
	}
	want = time.Now().Add(time.Hour)

	//缓存时间视为已偏移:Now() 的结果喂回 New,应原样返回,不再叠加 1h
	snap := Now()
	if got := New(snap).Now(); !got.Equal(snap) {
		t.Fatalf("缓存时间应原样返回: got %v, want %v", got, snap)
	}

	//派生 Times 继承偏差:缓存清零回落系统路径,偏差应仍在
	derived := Default.New(time.Time{})
	if d := derived.Now().Sub(want); d > 2*time.Second || d < -2*time.Second {
		t.Fatalf("派生 Times 未继承偏差: got %v, 期望≈%v, 差 %v", derived.Now(), want, d)
	}

	//偏差下届数不漂移:回拨两天后,"今天"仍应算第 0 届
	//若"现在"不走偏差,era(Daily(0) 按偏差日)与真实 now 相差 48h,会误报第 2 届
	if err := SetTimeOffset(-48 * time.Hour); err != nil {
		t.Fatalf("SetTimeOffset(-48h): %v", err)
	}
	if _, r := Default.Cycle(ExpireTypeDaily, 1).Cycle(); r != 0 {
		t.Fatalf("偏差 -48h 下 Daily Cycle=%d, 应为 0", r)
	}

	//偏差不追溯已构造的 Times:锚点在设偏差前创建(继承 offset=0),三天前绝对时刻仍报第 3 届
	if err := SetTimeOffset(time.Duration(0)); err != nil {
		t.Fatalf("SetTimeOffset(0): %v", err)
	}
	anchor := Default.Unix(time.Now().Unix() - 3*86400)
	if err := SetTimeOffset(48 * time.Hour); err != nil {
		t.Fatalf("SetTimeOffset(48h): %v", err)
	}
	if _, r := anchor.Cycle(ExpireTypeDaily, 1).Cycle(); r != 3 {
		t.Fatalf("绝对锚点 + 偏差 48h 下 Cycle=%d, 应为 3", r)
	}
}

// TestSince Since 按偏差后的"现在"求差:零偏差与 time.Since 同值,偏差下叠加偏差量
func TestSince(t *testing.T) {
	SetTimeOffset(0)
	begin := time.Now()
	time.Sleep(20 * time.Millisecond)
	if d := Since(begin); d < 20*time.Millisecond || d > 2*time.Second {
		t.Fatalf("零偏差 Since=%v, 期望 20ms~2s", d)
	}

	if err := SetTimeOffset(time.Hour); err != nil {
		t.Fatalf("SetTimeOffset(+1h): %v", err)
	}
	t.Cleanup(func() { SetTimeOffset(time.Duration(0)) })
	if d := Since(begin); d < time.Hour+20*time.Millisecond {
		t.Fatalf("+1h 偏差 Since=%v, 期望 ≥ 1h20ms", d)
	}
}
