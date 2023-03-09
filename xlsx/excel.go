package xlsx

import (
	"fmt"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/logger"
	"github.com/tealeg/xlsx/v3"
	"os"
	"path/filepath"
	"strings"
)

func LoadExcel(dir string) {
	var sheets []*GameSheet
	filter := map[string]*GameSheet{}

	files := GetFiles(dir, Ignore)
	for _, file := range files {
		//wb, err := spreadsheet.Open(file)
		wb, err := xlsx.OpenFile(file)
		logger.Info("解析文件:%v", file)
		if err != nil {
			logger.Fatal("excel文件格式错误:%v\n%v", file, err)
		}
		for _, sheet := range wb.Sheets {
			if v := ParseSheet(sheet); v != nil {
				if i, ok := filter[v.LowerName]; ok {
					logger.Warn("表格名字[%v]重复自动跳过\n----FROM:%v\n----TODO:%v", v.ProtoName, i.FileName, file)
				} else {
					v.FileName = file
					filter[v.LowerName] = v
					sheets = append(sheets, v)
				}
			}
		}
	}

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
	ProtoGo()
}

func CreateSheet(sheet *xlsx.Sheet) (skip int, row *GameSheet) {
	if !Valid(sheet) {
		return 0, nil
	}
	row = &GameSheet{SheetName: sheet.Name}
	max := sheet.MaxRow
	for skip = 0; skip <= max; skip++ {
		r, e := sheet.Row(skip)
		if e != nil {
			logger.Fatal("获取sheet行错误 name:%v,err:%v", sheet.Name, e)
		}
		cell := r.GetCell(0)
		if cell.Value == "" {
			continue
		}
		if row.ProtoName == "" {
			row.ProtoName = cell.Value
			row.LowerName = strings.ToLower(cell.Value)
		} else if row.SheetType == nil {
			row.SheetType = map[int]string{}
			for j := 0; j <= sheet.MaxCol; j++ {
				if c := r.GetCell(j); c != nil && c.Value != "" {
					row.SheetType[j] = strings.TrimSpace(c.Value)
				}
			}
		} else if row.SheetFields == nil {
			var end bool
			var field = &Field{}
			for j := 0; j <= sheet.MaxCol; j++ {
				if end = field.Parse(r.GetCell(j), j, row.SheetType[j]); end {
					if field.Compile() {
						row.SheetFields = append(row.SheetFields, field)
					}
					field = &Field{}
				}
			}
			//v.SheetFields = map[string]any{} //todo
		} else {
			break
		}
	}
	if row.ProtoName == "" || strings.HasPrefix(row.SheetName, "~") || strings.HasPrefix(row.ProtoName, "~") {
		row = nil
	}
	return
}

func ParseSheet(sheet *xlsx.Sheet) *GameSheet {
	//countArr := []int{1, 101, 201, 301}
	max := sheet.MaxRow
	logger.Info("----开始读取表格[%v],共有%v行", sheet.Name, max)
	skip, ret := CreateSheet(sheet)
	if ret == nil {
		return ret
	}
	for i := skip + 1; i <= max; i++ {
		r, e := sheet.Row(i)
		if e != nil {
			logger.Info("%v,err:%v", i, e)
		}
		cell := r.GetCell(0)
		if cell.Value != "" {
			//logger.Info("cell:%v  type:%v", cell.Value, cell.Type())
		}
	}

	//err := sheet.ForEachRow(func(r *xlsx.Row) error {
	//	cell := r.GetCell(0)
	//	if cell.Value == "" {
	//		return nil
	//	}
	//
	//	logger.Info("cell:%v  type:%v", cell.Value, cell.Type())
	//	return nil
	//	//return r.ForEachCell(func(c *xlsx.Cell) error {
	//	//	logger.Info("cell:%v", c.Value)
	//	//	return nil
	//	//})
	//})
	//if err != nil {
	//	logger.Info("err:%v", err)
	//}

	return ret
}

//func genProto2(gamedata []GameSheet) []byte {
//	sw := bytes.NewBufferString("")
//	t := template.New("")
//	t.Funcs(template.FuncMap{
//		"genRequire": func(t string) string {
//			switch t {
//			case "int":
//				return "required"
//			case "int64":
//				return "required"
//			case "str", "string":
//				return "required"
//			case "ti":
//				return "repeated"
//			case "ts":
//				return "repeated"
//			}
//			return ""
//		},
//		"genType": func(t string) string {
//			switch t {
//			case "int":
//				return "int32"
//			case "int64":
//				return "int64"
//			case "str", "string":
//				return "string"
//			case "ti":
//				return "int32"
//			case "ts":
//				return "string"
//			}
//			return ""
//		},
//	})
//	t.Parse(proto2Temple)
//	data := &struct {
//		Protos []GameSheet
//	}{
//		Protos: gamedata,
//	}
//	err := t.Execute(sw, data)
//	if err != nil {
//		fmt.Println(err)
//	}
//	//fmt.Println(sw.String())
//	return sw.Bytes()
//}

//
//func getCellString(wb *spreadsheet.Workbook, cell spreadsheet.Cell) string {
//	s := cell.GetString()
//	if s != "" {
//		return s
//	}
//	x := cell.X()
//	switch x.TAttr {
//	case sml.ST_CellTypeInlineStr:
//		if x.Is != nil && x.Is.T != nil {
//			return *x.Is.T
//		}
//		if x.V != nil {
//			return *x.V
//		}
//	case sml.ST_CellTypeS:
//		if x.V == nil {
//			return ""
//		}
//		id, err := strconv.Atoi(*x.V)
//		if err != nil {
//			return ""
//		}
//		if id < 0 {
//			return ""
//		}
//		if id > len(wb.SharedStrings.X().Si) {
//			return ""
//		}
//		si := wb.SharedStrings.X().Si[id]
//		if si.T != nil {
//			return *si.T
//		}
//		if len(si.R) > 0 {
//			rs := ""
//			for _, r := range si.R {
//				rs += r.T
//			}
//			return rs
//		}
//		return ""
//	}
//	if x.V == nil {
//		return ""
//	}
//	return *x.V
//}
