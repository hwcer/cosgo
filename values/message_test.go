package values

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestMessage_Error(t *testing.T) {

	msg := Message{}

	_ = msg.Parse(fmt.Errorf("test error"))
	t.Log(msg.String())

	_ = msg.Parse("test string")
	t.Log(msg.String())

	_ = msg.Parse(100)
	t.Log(msg.String())
	msg.Code = 0

	v := map[string]interface{}{}
	v["k"] = "k"
	v["v"] = 1
	msg.Data = v

	b, err := json.Marshal(msg)
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("test json Marshal:%s", string(b))
	}

	//模拟通过NET获得的message
	r := &Request{}
	if err = json.Unmarshal(b, r); err != nil {
		t.Error(err)
	}

	m := map[string]interface{}{}
	if err = r.Unmarshal(&m); err != nil {
		t.Error(err)
	} else {
		t.Logf("test net Unmarshal:%v", m)
	}

}
