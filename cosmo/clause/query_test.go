package clause

import (
	"testing"
)

func TestQuery(t *testing.T) {
	query := New()
	query.Where("_id = ?", 130)
	query.Where("_id IN ? AND uid = ?", []int{110, 120}, "myUid")

	query.Where("_id = ? OR uid = ?", 100, "GM")

	query.Where("_id != ?", 180)

	t.Logf("%v", query.String())
}
