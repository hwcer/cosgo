package app

import (
	"context"
	"fmt"
	"icefire/ftime"
	"icefire/logger"
	"icefire/sysutil/shm"
	"strings"
	"time"
)

func init() {
	checkStop = func(s context.CancelFunc, pid int) {
		tk := ftime.NewTicker(time.Second * 2)

		msize := int32(1)
		sstop := "1"
		rbuf := make([]byte, msize)
		shmName := fmt.Sprintf("shm.%v", pid)
		w, err := shm.Create(shmName, msize)
		if err != nil {
			logger.ERR("shm create failed:%v,%v", shmName, err)
			return
		}
		_, err = w.Write([]byte("0"))
		if err != nil {
			logger.ERR("shm init failed:%v,%v", shmName, err)
			return
		}

		for {
			select {
			case <-tk.C:
				r, err := shm.Open(shmName, msize)
				if err != nil {
					logger.ERR("shm open failed:%v,%v", shmName, err)
					return
				}
				defer r.Close()
				_, err = r.Read(rbuf)
				if err != nil {
					logger.ERR("shm read err:%v,%v", shmName, err)
					return
				}
				rsig := strings.TrimSpace(string(rbuf))
				if rsig == sstop {
					logger.INFO("shm stop")
					s()
					return
				}
			}
		}
	}
}
