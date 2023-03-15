package xlsx

import (
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/logger"
	"github.com/tealeg/xlsx/v3"
	"strings"
)

func LoadExcel(dir string) {
	logger.Info("====================开始解析静态数据====================")
	filter := map[string]*Message{}
	sheets := []*Message{}
	files := GetFiles(dir, Ignore)
	var protoIndex int
	for _, file := range files {
		//wb, err := spreadsheet.Open(file)
		wb, err := xlsx.OpenFile(file)
		logger.Info("解析文件:%v", file)
		if err != nil {
			logger.Fatal("excel文件格式错误:%v\n%v", file, err)
		}
		for _, sheet := range wb.Sheets {
			protoIndex += 1
			if v := ParseSheet(sheet, protoIndex); v != nil {
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
	writeExcelIndex(sheets)
	writeProtoMessage(sheets)
	if cosgo.Config.GetString(FlagsNameJson) != "" {
		writeValueJson(sheets)
	}
	if cosgo.Config.GetString(FlagsNameGo) != "" {
		ProtoGo()
	}

}

func CreateSheet(sheet *xlsx.Sheet) (row *Message) {
	if !Valid(sheet) {
		return nil
	}
	var skip int
	row = &Message{SheetName: sheet.Name}
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
			if c := r.GetCell(1); c != nil {
				switch strings.ToLower(c.Value) {
				case "kv":
					row.ExportType = ExportTypeKVS
				case "array":
					row.ExportType = ExportTypeARR
				}
			}
		} else if row.SheetType == nil {
			row.SheetType = map[int]string{}
			for j := 0; j <= sheet.MaxCol; j++ {
				if c := r.GetCell(j); c != nil && c.Value != "" {
					row.SheetType[j] = FormatType(strings.TrimSpace(c.Value))
				}
			}
		} else if row.Fields == nil {
			var end bool
			var field = &Field{}
			for j := 0; j <= sheet.MaxCol; j++ {
				if end = field.Parse(r.GetCell(j), j, row.SheetType[j]); end {
					if field.Compile() {
						row.Fields = append(row.Fields, field)
					}
					field = &Field{}
				}
			}
			//v.Fields = map[string]any{} //todo
		} else if row.SheetDesc == nil {
			row.SheetDesc = map[int]string{}
			for j := 0; j <= sheet.MaxCol; j++ {
				if c := r.GetCell(j); c != nil {
					row.SheetDesc[j] = c.Value
				}
			}
			break
		}
	}
	if row.ProtoName == "" || strings.HasPrefix(row.SheetName, "~") || strings.HasPrefix(row.ProtoName, "~") {
		return nil
	}

	row.SheetSkip = skip
	row.SheetRows = sheet
	for _, field := range row.Fields {
		if field.ProtoRequire == FieldTypeNone {
			field.ProtoDesc = strings.ReplaceAll(row.SheetDesc[field.Index[0]], "\n", "")
		} else {
			field.ProtoDesc = field.ProtoName
		}
	}
	return
}

func ParseSheet(sheet *xlsx.Sheet, index int) (r *Message) {
	//countArr := []int{1, 101, 201, 301}
	max := sheet.MaxRow
	logger.Info("----开始读取表格[%v],共有%v行", sheet.Name, max)
	r = CreateSheet(sheet)
	if r != nil {
		r.ProtoIndex = index
	}
	return
}
