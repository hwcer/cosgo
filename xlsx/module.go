package xlsx

import (
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/logger"
)

const (
	FlagsNameIn     string = "in"
	FlagsNameGo     string = "go"
	FlagsNameOut    string = "out"
	FlagsNameJson   string = "json"
	FlagsNameIgnore string = "ignore"
)

var mod *module

func init() {
	cosgo.Config.Flags(FlagsNameIn, "", "./", "需要解析的excel目录")
	cosgo.Config.Flags(FlagsNameOut, "", "./data", "静态数据文件目录,包括data.pb data.json") // message data.pb   data.json
	cosgo.Config.Flags(FlagsNameJson, "", true, "是否导json格式")
	cosgo.Config.Flags(FlagsNameGo, "", "./data", "生成的GO文件")
	cosgo.Config.Flags(FlagsNameIgnore, "", "", "忽略的文件或者文件夹逗号分割多个")
}

func New() *module {
	if mod == nil {
		mod = &module{}
	}
	return mod
}

type module struct {
	cosgo.Module
}

func (this *module) Start() error {
	//CdProgramDir()
	//Init()
	preparePath()
	LoadExcel(cosgo.Config.GetString(FlagsNameIn))
	logger.Info("\n========================恭喜大表哥导表成功========================\n")
	return nil
	//
	//if flagWatch {
	//	watcher, err := fsnotify.NewWatcher()
	//	if err != nil {
	//		color.Red.Println("NewWatcher failed: ", err)
	//		return
	//	}
	//	defer watcher.Close()
	//	err = recursiveWatch(watcher, flagExcelPath)
	//	if err != nil {
	//		color.Red.Println("AddExcel Path failed: ", err)
	//		return
	//	}
	//	go func() {
	//		fn := NewDebouncer(time.Second)
	//		for value := range watcher.Events {
	//			fmt.Println(value.String())
	//			if !strings.HasSuffix(value.Name, ".xlsx") {
	//				continue
	//			}
	//			fn(func() {
	//				err := generateFiles()
	//				if err != nil {
	//					color.Red.Println(err)
	//					beeep.Notify("excel解析出错!", err.Error(), "assets/warning.png")
	//				}
	//			})
	//		}
	//	}()
	//	select {}
	//}
}
