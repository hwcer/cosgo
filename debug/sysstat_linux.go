package debug

import (
	"log"
	"syscall"
	"time"
)

func showSystemStat(interval time.Duration, count int) {

	usage1 := &syscall.Rusage{}
	var lastUtime int64
	var lastStime int64

	counter := 0
	for {
		//http://man7.org/linux/man-pages/man3/vtimes.3.html
		syscall.Getrusage(syscall.RUSAGE_SELF, usage1)

		utime := usage1.Utime.Nano()
		stime := usage1.Stime.Nano()
		userCPUUtil := float64(utime-lastUtime) * 100 / float64(interval)
		sysCPUUtil := float64(stime-lastStime) * 100 / float64(interval)
		memUtil := usage1.Maxrss * 1024

		lastUtime = utime
		lastStime = stime

		if counter > 0 {
			log.Printf("cpu: %3.2f%% us  %3.2f%% sy, mem:%s\n", userCPUUtil, sysCPUUtil, toH(uint64(memUtil)))
		}

		counter += 1
		if count >= 1 && count < counter {
			return
		}
		time.Sleep(interval)
	}
}
