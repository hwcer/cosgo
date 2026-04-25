package main

import (
	"fmt"
	"reflect"

	"github.com/hwcer/cosgo/schema"
)

// Paging 模拟 cosmo.Paging 结构体
type Paging struct {
	Page     int `json:"page" bson:"page"`
	PageSize int `json:"pageSize" bson:"pageSize"`
}

func main() {
	// 模拟用户代码中的临时结构体
	args := struct {
		Paging
		Iid int32 `json:"Iid" bson:"Iid"`
	}{}

	// 尝试解析这个临时结构体
	sch, err := schema.GetOrParse(&args, nil)
	if err != nil {
		fmt.Printf("Error parsing schema: %v\n", err)
		return
	}
	if sch == nil {
		fmt.Println("Error: schema is nil")
		return
	}

	// 打印schema信息
	fmt.Printf("Schema: %v\n", sch)
	fmt.Printf("Schema.Fields: %v\n", sch.Fields)
	fmt.Printf("Schema.Fields == nil: %v\n", sch.Fields == nil)
	fmt.Printf("Schema.Embedded: %v\n", sch.Embedded)

	// 尝试访问字段
	fmt.Println("\nSchema fields:")
	if sch.Fields != nil {
		for name, field := range sch.Fields {
			fmt.Printf("Field: %s, Index: %v\n", name, field.Index)
		}
	} else {
		fmt.Println("schema.Fields is nil")
	}

	// 尝试获取值
	args.Iid = 123
	args.Page = 1
	args.PageSize = 10

	fmt.Println("\nField values:")
	for name, field := range sch.Fields {
		value := field.Get(reflect.ValueOf(&args))
		fmt.Printf("Field: %s, Value: %v\n", name, value.Interface())
	}
}
