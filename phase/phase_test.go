package phase

import "testing"

// 阶段推进与守卫判定:Sealed 在 Started 及之后恒真,关闭期注册同样必须被拒
func TestSealed(t *testing.T) {
	defer Set(Init)
	cases := []struct {
		p      Type
		sealed bool
	}{
		{Init, false},
		{Starting, false},
		{Started, true},
		{Closing, true},
		{Closed, true},
	}
	for _, c := range cases {
		Set(c.p)
		if got := Sealed(); got != c.sealed {
			t.Fatalf("phase=%v Sealed()=%v, want %v", c.p, got, c.sealed)
		}
		if !Is(c.p) {
			t.Fatalf("Is(%v) 应为真", c.p)
		}
	}
}

func TestString(t *testing.T) {
	if Started.String() != "Started" {
		t.Fatalf("Started.String()=%v", Started)
	}
	if s := Type(99).String(); s != "Phase(99)" {
		t.Fatalf("未知阶段应保留数值:%v", s)
	}
}
