package values

import (
	"encoding/json"
	"fmt"
	"testing"
)

type items struct {
	Attach Attach
}

func TestAttach_Marshal(t *testing.T) {
	i := &items{}
	_ = i.Attach.Marshal("ok")

	log(i)

	i = &items{}
	_ = i.Attach.Marshal(23)
	log(i)
	i = &items{}
	_ = i.Attach.Marshal(map[int32]int32{1: 12})

	log(i)
}

func log(i any) {
	b, _ := json.Marshal(i)
	fmt.Printf("----:%v\n", string(b))
}
