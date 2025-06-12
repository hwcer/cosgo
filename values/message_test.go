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
	t.Logf("test map value:%v", v)
	_ = msg.Marshal(v)

	v2 := map[string]interface{}{}
	if err := msg.Unmarshal(&v2); err != nil {
		t.Error(err)
	} else {
		t.Logf("test json Unmarshal:%v", v2)
	}

	b, err := msg.MarshalJSON()
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("test json Marshal:%s", string(b))
	}

	//模拟通过NET获得的message
	m2 := &Message{}
	if err = json.Unmarshal(b, m2); err != nil {
		t.Error(err)
	}

	r := map[string]interface{}{}
	if err = m2.Unmarshal(&r); err != nil {
		t.Error(err)
	} else {
		t.Logf("test net Unmarshal:%v", r)
	}

}
