package schema

import (
	"encoding/json"
	"testing"
)

type Sex struct {
	Hair int `json:"hair" json:"hair"`
}

type Human struct {
	Sex  Sex    `bson:"sex" json:"sex"`
	Age  int32  `json:"age" bson:"age"`
	Name string `json:"name" bson:"name"`
}

func TestSchema_GetValue(t *testing.T) {
	h := &Human{Sex: Sex{Hair: 1}}

	sch, _ := Parse(h)

	err := sch.SetValue(h, "sex.hair", 2)
	if err != nil {
		t.Logf("error:%v", err)
	} else {
		b, _ := json.Marshal(h)
		t.Logf("result:%v", string(b))
	}

}
