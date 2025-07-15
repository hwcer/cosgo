package values

import (
	"encoding/json"
	"testing"
)

func TestValues(t *testing.T) {
	vs := Values{}
	vs.Set("a", "a")
	vs.Set("b", true)
	vs.Set("c", 1)
	_, _ = vs.Marshal("d", "d")
	t.Logf("a:%s", vs.GetString("a"))
	t.Logf("b:%v", vs.Get("b"))
	t.Logf("c:%d", vs.GetInt("c"))

	_, _ = vs.Marshal("arr", []int{1, 2, 3})

	var arr []int
	if err := vs.Unmarshal("arr", &arr); err != nil {
		t.Logf("%v", err)
	} else {
		t.Logf("arr:%v", arr)
	}

	if b, err := json.Marshal(vs); err != nil {
		t.Logf("%v", err)
	} else {
		t.Logf("%s", string(b))
	}

}
