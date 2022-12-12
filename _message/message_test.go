package message

import (
	"encoding/json"
	"errors"
	"testing"
)

type User struct {
	Name string
	Icon int
}

func TestMessage(t *testing.T) {
	u := &User{Name: "test", Icon: 100}
	r := Parse(u)
	b, err := json.Marshal(r)
	if err != nil {
		t.Logf(err.Error())
	} else {
		t.Logf(string(b))
	}

	u2 := &User{}
	err = r.Unmarshal(u2)
	if err != nil {
		t.Logf(err.Error())
	} else {
		t.Logf("%+v", *u2)
	}

	r = &Message{}
	b, err = json.Marshal(r)
	if err != nil {
		t.Logf(err.Error())
	} else {
		t.Logf(string(b))
	}

	r = Parse(errors.New("test error"))
	b, _ = json.Marshal(r)
	t.Logf(string(b))

	r = Parse("test string")
	b, _ = json.Marshal(r)
	t.Logf(string(b))

}
