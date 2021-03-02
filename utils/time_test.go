package utils

import (
	"testing"
	"time"
)

func TestGetDayStartTime(t *testing.T) {
	d := time.Now()
	s := GetDayTime(d, 0)
	t.Logf("%v", s)

}
func TestGetTimeWeekStart(t *testing.T) {
	d := time.Now()
	for i := 0; i < 10; i++ {
		d2 := d.AddDate(0, 0, -i)
		s := GetWeekTime(d2, 0)
		t.Logf("%v ==> %v", d2, s)
	}
}
