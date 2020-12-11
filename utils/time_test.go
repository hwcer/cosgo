package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestGetDayStartTime(t *testing.T) {

	d := GetTimeDayStart(time.Now(), 1, 6)
	fmt.Printf("时间:%v\n", d.Format(time.RFC3339))
}
