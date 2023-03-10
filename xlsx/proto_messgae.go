package xlsx

import (
	"fmt"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/logger"
	"os"
	"path/filepath"
	"strings"
)

func writeProtoMessage(sheets []*Message) {
	logger.Info("======================开始生成PROTO MESSAGE======================")
	//生成proto
	b := &strings.Builder{}
	ProtoTitle(b)
	//输出所有标签
	b.WriteString("\n//配置索引......\n")
	in := cosgo.Config.GetString(FlagsNameIn) + "/"
	for _, s := range sheets {
		b.WriteString(fmt.Sprintf("//[%v]%v:%v\n", s.ProtoName, s.SheetName, strings.TrimPrefix(s.FileName, in)))
	}

	b.WriteString("\n//全局对象......\n")
	buildGlobalObjects(b, sheets)
	b.WriteString("\n//数据对象......\n")
	ProtoMessage(sheets, b)
	file := filepath.Join(cosgo.Config.GetString(FlagsNameOut), "message.proto")
	if err := os.WriteFile(file, []byte(b.String()), os.ModePerm); err != nil {
		logger.Fatal(err)
	}
	logger.Info("Proto Message File:%v", file)
}
