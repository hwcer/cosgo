package schema

import (
	"encoding/json"
	"testing"
)

type Sex struct {
	Hair int `json:"hair" json:"hair"`
}

type A struct {
	X string
	B
}

type B struct {
	Y int
}

type Human struct {
	A
	Sex  *Sex           `bson:"sex" json:"sex"`
	Age  int32          `json:"age" bson:"age"`
	Name string         `json:"name" bson:"name"`
	Args map[string]any `json:"args" bson:"args"`
}

func TestSchema_GetValue(t *testing.T) {
	h := &Human{Sex: &Sex{}}
	h.Y = 2
	h.Args = make(map[string]any)
	sch, _ := Parse(h)
	err := sch.SetValue(h, 2, "sex", "hair")
	if err != nil {
		t.Logf("error:%v", err)
	} else {
		b, _ := json.Marshal(h)
		t.Logf("result:%v", string(b))
	}

	err = sch.SetValue(h, 200, "Y")
	if err != nil {
		t.Logf("error:%v", err)
	} else {
		b, _ := json.Marshal(h)
		t.Logf("result:%v", string(b))
	}

	err = sch.SetValue(h, 200, "Args", "val")
	if err != nil {
		t.Logf("error:%v", err)
	} else {
		b, _ := json.Marshal(h)
		t.Logf("result:%v", string(b))
	}

	t.Logf("get sex.Hair:%v", sch.GetValue(h, "sex", "hair"))
}
