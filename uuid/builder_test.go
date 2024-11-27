package uuid

import (
	"testing"
)

func TestUUID(t *testing.T) {
	b := New(1, 1000)

	uid := b.New(0)
	t.Logf("uid[10]:%v\n", uid.String(10))
	t.Logf("uid[32]:%v\n", uid.String(32))

	x := uid.String(10)
	s, p, _ := Split(x, 10, 2)
	t.Logf("uid[10] Split:%v--%v\n", s, p)

	u2 := &UUID{}
	if err := u2.Parse(uid.String(10), 10); err != nil {
		t.Error(err)
	} else {
		t.Logf("uid2[10]:%v\n", u2.String(10))
	}

	u3 := &UUID{}
	if err := u3.Parse(uid.String(32), 32); err != nil {
		t.Error(err)
	} else {
		t.Logf("uid3[32]:%v\n", u3.String(10))
	}
}
