package values

import (
	"encoding/json"
	"github.com/hwcer/cosgo/bson"
	"testing"
)

type data struct {
	Id  int32
	Msg Bytes `bson:"msg"`
}

func TestBytes_JSON(t *testing.T) {
	i := &data{Id: 111, Msg: nil}
	//_ = i.Msg.Marshal("null")
	b, err := json.Marshal(i)
	if err != nil {
		t.Logf("ERROR:%v", err)
	} else {
		t.Logf("BYTES:%v", b)
		t.Logf("SUCCESS:%v", string(b))
	}

	j := &data{}
	err = json.Unmarshal(b, j)
	if err != nil {
		t.Logf("ERROR:%v", err)
	} else {
		t.Logf("SUCCESS:%+v   MSG:%v", j, len(j.Msg))
	}

}
func TestBytes_BSON(t *testing.T) {
	i := &data{Id: 111, Msg: nil}
	//_ = i.Msg.Marshal("null")
	b, err := bson.Marshal(i)
	if err != nil {
		t.Logf("ERROR:%v", err)
		return
	}
	t.Logf("BYTES:%v", b)
	t.Logf("VALUE:%v", b.Value("msg").Data)

	j := &data{}
	err = bson.Unmarshal(b, j)
	if err != nil {
		t.Logf("ERROR:%v", err)
	} else {
		t.Logf("SUCCESS:%+v   MSG:%v", j, len(j.Msg))
	}

}
