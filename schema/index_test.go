package schema

import (
	"fmt"
	"testing"
)

var opts = New()

type Member struct {
	Id    string `json:"_id" bson:"_id" ` //user id
	Icon1 string `index:"name:idx_icon"`
	Icon2 string `index:"name:idx_icon"`
	Name  string `json:"name,omitempty" bson:"name" index:"name:"`
	Arena Arena  `json:"arena,omitempty" bson:"arena"`
}

type Arena struct {
	Score  int32 `json:"score" bson:"score" index:"name:"`
	Circle Circle
}
type Circle struct {
	Rank int32 `index:"name:"`
}

func TestIndex(t *testing.T) {
	role := &Member{}
	sch, _ := opts.Parse(role)
	indexes := sch.ParseIndexes()
	fmt.Printf("------------------\n")
	for k, v := range indexes {
		i := v.Build()
		fmt.Printf("index:%v,keys:%+v\n", k, i.Keys)
	}

}
