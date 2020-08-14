package debug

import (
	"golang.org/x/sys/windows"
	"log"
	"syscall"
	"time"
	"unsafe"
)

// https://github.com/shirou/gopsutil
// https://github.com/shirou/gopsutil/blob/master/process/process_windows.go

var (
	modpsapi                 = windows.NewLazySystemDLL("psapi.dll")
	procGetProcessMemoryInfo = modpsapi.NewProc("GetProcessMemoryInfo")
)

type PROCESS_MEMORY_COUNTERS struct {
	CB                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uint64 //峰值内存使用
	WorkingSetSize             uint64 //内存使用
	QuotaPeakPagedPoolUsage    uint64
	QuotaPagedPoolUsage        uint64
	QuotaPeakNonPagedPoolUsage uint64
	QuotaNonPagedPoolUsage     uint64
	PagefileUsage              uint64 //虚拟内存使用
	PeakPagefileUsage          uint64 //峰值虚拟内存使用
}

func getMemoryInfo(pid int32) (PROCESS_MEMORY_COUNTERS, error) {
	var mem PROCESS_MEMORY_COUNTERS
	c, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return mem, err
	}
	defer windows.CloseHandle(c)
	if err := getProcessMemoryInfo(c, &mem); err != nil {
		return mem, err
	}

	return mem, err
}

func getProcessMemoryInfo(h windows.Handle, mem *PROCESS_MEMORY_COUNTERS) (err error) {
	r1, _, e1 := syscall.Syscall(procGetProcessMemoryInfo.Addr(), 3, uintptr(h), uintptr(unsafe.Pointer(mem)), uintptr(unsafe.Sizeof(*mem)))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func showSystemStat(interval time.Duration, count int) {

	var usage1 windows.Rusage
	var lastUtime int64
	var lastStime int64

	counter := 0
	for {
		c, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, uint32(pid))
		if err != nil {
			log.Printf("open process failed:%v\n", err)
			return
		}
		defer windows.CloseHandle(c)

		if err := windows.GetProcessTimes(c, &usage1.CreationTime, &usage1.ExitTime, &usage1.KernelTime, &usage1.UserTime); err != nil {
			log.Printf("get cpu usage failed:%v\n", err)
			return
		}

		mem, err := getMemoryInfo(int32(pid))
		if err != nil {
			log.Printf("get mem usage failed:%v\n", err)
			return
		}

		utime := usage1.UserTime.Nanoseconds()
		stime := usage1.KernelTime.Nanoseconds()
		userCPUUtil := float64(utime-lastUtime) * 100 / float64(interval)
		sysCPUUtil := float64(stime-lastStime) * 100 / float64(interval)
		memUtil := mem.WorkingSetSize

		lastUtime = utime
		lastStime = stime

		if counter > 0 {
			log.Printf("cpu: %3.2f%% us  %3.2f%% sy, mem:%s \n", userCPUUtil, sysCPUUtil, toH(uint64(memUtil)))
		}

		counter += 1
		if count >= 1 && count < counter {
			return
		}
		time.Sleep(interval)
	}

	/*
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
				fmt.Printf("cpu: %3.2f%% us  %3.2f%% sy, mem:%s \n", userCPUUtil, sysCPUUtil, toH(uint64(memUtil)))
			}

			counter += 1
			if count >= 1 && count < counter {
				return
			}
			time.Sleep(interval)
		}
	*/
}
