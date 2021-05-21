package message

import (
	"testing"
)

func init() {
	SetAttachField(1, "b", 4)
}

func TestMessage(t *testing.T) {
	m := New(2, 1, ContentTypeNumber)
	m.Attach.Set("a", 1)
	m.Attach.Set("c", 3)
	t.Logf("Message:%+v,%v\n", m.Head, m.Attach)

	b := m.Head.Bytes()
	t.Logf("head len:%v, byte:%v", len(b), b)

	msg, err := NewMsg(b)
	if err != nil {
		t.Logf("new head err %v:", err)
	} else {
		t.Logf("new head %+v", msg.Head)
	}

	t.Logf("flag has %v:%b", msg.Head.Flags.Has(MsgFlagCompress), msg.Head.Flags)
	msg.Head.Flags.Add(MsgFlagCompress)
	t.Logf("flag has %v:%b", msg.Head.Flags.Has(MsgFlagCompress), msg.Head.Flags)
	msg.Head.Flags.Del(MsgFlagCompress)
	t.Logf("flag has %v:%b", msg.Head.Flags.Has(MsgFlagCompress), msg.Head.Flags)

	a := map[int]int{1: 1, 2: 2, 3: 3, 4: 4, 5: 5}
	for k, v := range a {
		t.Logf("%v:%v", k, v)
		delete(a, k)
	}

}
