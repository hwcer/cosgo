package xlsx

import (
	"fmt"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/logger"
)

func generateFiles() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%w", e)
		}
	}()
	LoadExcel(cosgo.Config.GetString(FlagsNameIn))
	//if flagParseGo {
	//	files, _ := ioutil.ReadDir(flagProtoPath)
	//	for _, filename := range files {
	//		if strings.HasSuffix(filename.Name(), ".proto") {
	//			err := os.Remove(path.Join(flagProtoPath, filename.Name()))
	//			if err != nil {
	//				panic(err.Error())
	//			}
	//		}
	//	}
	//	proto := genProto2(gamedata)
	//	ioutil.WriteFile(path.Join(flagProtoPath, "data.proto"), proto, os.ModePerm)
	//	if !flagParseJson {
	//		genProtoGo()
	//	}
	//}
	//if !flagParseJson {
	//	genGendata2(gamedata)
	//} else {
	//	genJsonData(gamedata)
	//}
	logger.Info("~~~~~~~~~~恭喜大表哥导表成功~~~~~~~~~~")
	return
}

//
//func genGendata2(gamedata []GameSheet) {
//
//	for _, sheet := range gamedata {
//
//		bf := bytes.NewBuffer(nil)
//		binary.Write(bf, byteOrder, byte(1))
//		binary.Write(bf, byteOrder, byte(1))
//		binary.Write(bf, byteOrder, uint32(len(sheet.Rows)))
//
//		for row, _ := range sheet.Rows {
//			pb := &AnyPb{}
//			rv := reflect.ValueOf(pb)
//			for idx, meta := range sheet.RowMetas {
//				value := sheet.Rows[row][idx].Value
//				switch meta.Type {
//				case "int":
//					field := rv.Elem().FieldByName("Int" + strconv.Itoa(meta.IndexVersion2))
//					field.Set(reflect.ValueOf(proto.Int32(value.(int32))))
//				case "str", "string":
//					field := rv.Elem().FieldByName("Str" + strconv.Itoa(meta.IndexVersion2-100))
//					field.Set(reflect.ValueOf(proto.String(value.(string))))
//				case "ti":
//					if value != nil {
//						field := rv.Elem().FieldByName("Intarr" + strconv.Itoa(meta.IndexVersion2-200))
//						field.Set(reflect.ValueOf(value.([]int32)))
//					}
//				case "ts":
//					if value != nil {
//						field := rv.Elem().FieldByName("Strarr" + strconv.Itoa(meta.IndexVersion2-300))
//						field.Set(reflect.ValueOf(value.([]string)))
//					}
//				}
//			}
//			pbBytes, err := proto.Marshal(pb)
//			if err != nil {
//				panic(err)
//			}
//			binary.Write(bf, byteOrder, uint32(len(pbBytes)))
//			bf.Write(pbBytes)
//		}
//		ioutil.WriteFile(path.Join(flagPbBytesPath, sheet.SheetName+".protodata.bytes"), bf.Bytes(), os.ModePerm)
//	}
//}
//
//func genJsonData(gameSheets []GameSheet) {
//	var arr []map[string]interface{}
//	for _, sheet := range gameSheets {
//		// sheet.SheetName + ".json"
//		for _, value := range sheet.Rows {
//			m := map[string]interface{}{}
//			for idx, meta := range sheet.RowMetas {
//				m[meta.Key] = value[idx].Value
//			}
//			arr = append(arr, m)
//		}
//		bs, _ := json.Marshal(arr)
//		arr = nil
//		ioutil.WriteFile(path.Join(flagJsonPath, sheet.SheetName+".json"), bs, os.ModePerm)
//	}
//}
