package mongo

import "testing"

type kv struct {
	K string `bson:"k"`
	V int32
}

func TestNewUpdate(t *testing.T) {
	update := make(Update)
	update.Set("set", 1)
	update.Inc("inc1", 21)
	update.Inc("inc2", 22)
	update.Max("max", 3)
	update.Unset("unset1", "who1")
	update.Unset("unset2", "who2")
	t.Logf("update:%v", update)

	u2 := make(Update)
	u2.Max("max", 3)
	t.Logf("update:%v", u2)

	u3 := make(Update)
	t.Logf("update:%v", u3)

	m := make(map[string]interface{})
	m["name"] = "test"
	m["value"] = 1

	u4 := make(Update)

	t.Logf("update:%v", u4)

}
