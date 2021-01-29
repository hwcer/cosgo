package cosnet

import (
	"testing"
)

func TestMessage(t *testing.T) {
	head := &Header{
		Size:     10,
		Index:    254,
		Proto:    2,
		DataType: 3,
		Flags:    MsgFlagType(100),
	}

	b := head.Bytes()
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
