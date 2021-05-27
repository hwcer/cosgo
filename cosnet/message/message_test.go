package message

import (
	"testing"
)

func init() {
	SetAttachField("a", 4)
	SetAttachField("b", 4)
}

func TestMessage(t *testing.T) {
	m := New(2, int32(1), ContentTypeNumber)
	err := m.Head.Attach.Set("a", 100)
	if err != nil {
		t.Logf("Err:%v\n", err)
	}
	err = m.Head.Attach.Set("b", 350)
	if err != nil {
		t.Logf("Err:%v\n", err)
	}
	t.Logf("Message:%+v,%v\n", m.Head, m.Head.Attach)

	b := m.Head.Bytes()
	t.Logf("head len:%v, byte:%v", len(b), b)

	attach := m.Head.Attach.Bytes()
	t.Logf("AttachIndex len:%v, byte:%v", len(attach), attach)

	msg, _ := NewMsg(b)

	t.Logf("new head %+v", msg.Head)
	msg.Head.Attach.Parse(attach)
	t.Logf("new Attach %+v", msg.Head.Attach)

	var a int32
	msg.Head.Attach.Get("a", &a)
	t.Logf("new Attach val a:%+v", a)
}
