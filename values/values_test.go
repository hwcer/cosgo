package values

import "testing"

func TestValues(t *testing.T) {
	vs := Values{}
	vs.Set("a", "a")
	vs.Set("b", true)
	vs.Set("c", 1)
	_, _ = vs.Marshal("d", "d")
	_, _ = vs.Marshal("arr", []int{1, 2, 3})

	t.Logf("a:%s", vs.GetString("a"))
	t.Logf("b:%v", vs.Get("b"))
	t.Logf("c:%d", vs.GetInt("c"))

	var d string
	if err := vs.Unmarshal("d", &d); err != nil {
		t.Logf("%v", err)
	}

	t.Logf("d:%s", d)

	var arr []int
	_ = vs.Unmarshal("arr", &arr)
	t.Logf("arr:%v", arr)

}
