package debug

import (
	"log"
	"testing"
	"time"
)

func TestDebug(t *testing.T) {
	StartPprofSrv(":7070")

	go func() {
		for i := 0; i < 300000000; i++ {
			a := 10000000.123 * 153468498.12 / 1234568.78
			a = a + a
			time.Sleep(time.Microsecond)
		}
	}()

	go func() {
		for i := 0; i < 300; i++ {
			log.Printf("GC Summory:%v", GCSummary())
			time.Sleep(time.Second * 5)
		}

	}()

	go func() {
		showSystemStat(time.Second*10, 5)
	}()
	time.Sleep(time.Second * 300)
}
