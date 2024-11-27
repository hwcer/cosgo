package schema

import (
	"encoding/json"
	"testing"
)

var opts = New()

type Member struct {
	Id      string `json:"_id" bson:"_id" `
	Name    string `json:"name,omitempty" bson:"name" index:"name:"`
	Map     map[int32]int32
	Slice   []int32
	Array   [5]Circle
	Arena   Arena
	*Circle `bson:",inline"`
}

type Arena struct {
	Score int32 `json:"score" bson:"score" index:"name:"`
}
type Circle struct {
	Rank  int32 `json:"rank" bson:"rank" index:"name:"`
	Score int32
}

func TestIndex(t *testing.T) {
	role := &Member{}
	role.Circle = &Circle{Rank: 10}
	role.Arena.Score = 200
	role.Map = map[int32]int32{1: 1, 2: 2}
	role.Slice = []int32{1, 2, 3, 4, 5}
	role.Array = [5]Circle{{Rank: 0}, {Rank: 1}, {Rank: 2}}

	//v := reflect.ValueOf(role)
	sch, err := opts.Parse(role)
	if err != nil {
		t.Logf("sch err:%v", err)
	}
	if err = sch.SetValue(role, 3, "Rank"); err != nil {
		t.Logf("sch SetValue Circle.Rank:%v", err)
	}
	if err = sch.SetValue(role, 50000, "Score"); err != nil {
		t.Logf("sch SetValue Score:%v", err)
	}
	if err = sch.SetValue(role, 100, "Map", 1); err != nil {
		t.Logf("sch SetValue Map:%v", err)
	}

	t.Logf("GET Map.1:%+v", sch.GetValue(role, "Map", 1))
	t.Logf("GET Slice.1:%+v", sch.GetValue(role, "Slice", 1))
	t.Logf("GET Array.1.Rank:%+v", sch.GetValue(role, "Array", 1, "Rank"))
	t.Logf("GET Rank:%+v", sch.GetValue(role, "Rank"))
	t.Logf("GET Arena.Score:%+v", sch.GetValue(role, "Arena", "Score"))

	b, _ := json.Marshal(role)

	t.Logf("role:%+v", string(b))

	for k, v := range sch.ParseIndexes() {
		t.Logf("index[%v]:%+v", k, v)
	}

}
