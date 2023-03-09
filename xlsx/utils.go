package xlsx

import (
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/logger"
	"github.com/tealeg/xlsx/v3"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func Ignore(f string) bool {
	_, name := filepath.Split(f)
	if strings.HasPrefix(name, "~") {
		return false
	}
	if !strings.HasSuffix(f, ".xlsx") {
		return false
	}
	for _, v := range ignoreFiles {
		if strings.HasPrefix(f, v) {
			return false
		}
	}
	return true
}

func Valid(sheet *xlsx.Sheet) bool {
	r, e := sheet.Row(0)
	if e != nil {
		logger.Fatal("获取sheet行错误 name:%v,err:%v", sheet.Name, e)
	}
	cell := r.GetCell(0)
	return cell != nil && cell.Value != ""
}

func FirstUpper(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func GetFiles(dir string, filter func(string) bool) (r []string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		logger.Fatal(err)
	}
	for _, info := range files {
		if info.IsDir() {
			r = append(r, GetFiles(path.Join(dir, info.Name()), filter)...)
		} else {
			f := path.Join(dir, info.Name())
			if filter(f) {
				r = append(r, f)
			}
		}
	}
	return
}

func preparePath() {
	// excel文件必须存在
	logger.Info("====================开始检查EXCEL路径====================")
	root := cosgo.Dir()
	in := path.Join(root, cosgo.Config.GetString(FlagsNameIn))
	if excelStat, err := os.Stat(in); err != nil || !excelStat.IsDir() {
		logger.Fatal("excel路径必须存在且为目录: %v ", in)
	}
	cosgo.Config.Set(FlagsNameIn, in)

	logger.Info("====================开始检查pb输出路径====================")
	out := path.Join(root, cosgo.Config.GetString(FlagsNameOut))
	if excelStat, err := os.Stat(out); err != nil || !excelStat.IsDir() {
		logger.Fatal("静态数据目录错误: %v ", out)
	}
	files, _ := os.ReadDir(out)
	logger.Info("删除输出路径中的文件")
	for _, filename := range files {
		if strings.HasSuffix(filename.Name(), ".proto") ||
			strings.HasSuffix(filename.Name(), ".pb") ||
			strings.HasSuffix(filename.Name(), ".json") {
			err := os.Remove(path.Join(out, filename.Name()))
			if err != nil {
				logger.Fatal(err)
			}
		}
	}
	cosgo.Config.Set(FlagsNameOut, out)

	logger.Info("====================开始检查GO输出路径====================")
	if p := cosgo.Config.GetString(FlagsNameGo); p != "" {
		goOutPath := path.Join(root, p)
		if excelStat, err := os.Stat(goOutPath); err != nil || !excelStat.IsDir() {
			logger.Fatal("GO文件输出目录错误: %v ", goOutPath)
		}
		fs, _ := os.ReadDir(out)
		logger.Info("删除输出路径中的文件")
		for _, filename := range fs {
			if strings.HasSuffix(filename.Name(), ".go") {
				err := os.Remove(path.Join(out, filename.Name()))
				if err != nil {
					logger.Fatal(err)
				}
			}
		}
		cosgo.Config.Set(FlagsNameOut, out)
	}

	logger.Info("====================开始检查忽略文件列表====================")
	if s := cosgo.Config.GetString(FlagsNameIgnore); s != "" {
		for _, v := range strings.Split(s, ",") {
			if v != "" {
				ignoreFiles = append(ignoreFiles, path.Join(in, v))
			}
		}
	}

}
