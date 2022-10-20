package values

import (
	"encoding/json"
	"testing"
)

type data struct {
	Id  int32
	Msg Bytes
}

func TestBytes(t *testing.T) {
	i := &data{}

	b, err := json.Marshal(i)
	if err != nil {
		t.Logf("ERROR:%v", err)
	} else {
		t.Logf("SUCCESS:%v", string(b))
	}

	j := &data{}
	err = json.Unmarshal(b, j)
	if err != nil {
		t.Logf("ERROR:%v", err)
	} else {
		t.Logf("SUCCESS:%+v", j)
	}

}
