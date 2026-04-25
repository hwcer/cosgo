package main

import (
	"fmt"

	"github.com/hwcer/cosgo/schema"
)

// 测试用的嵌入结构体
type Address struct {
	City    string `json:"city" bson:"city"`
	Street  string `json:"street" bson:"street"`
	ZipCode string `json:"zipCode" bson:"zipCode"`
}

// 测试用的用户结构体
type User struct {
	Name    string  `json:"name" bson:"name"`
	Age     int     `json:"age" bson:"age"`
	Address Address `json:"address" bson:"address"`
}

// 测试用的订单结构体
type Order struct {
	ID     int     `json:"id" bson:"id"`
	UserID int     `json:"userId" bson:"userId"`
	Amount float64 `json:"amount" bson:"amount"`
}

// 测试用的复杂结构体，包含指针类型的嵌入字段
type Customer struct {
	ID     int    `json:"id" bson:"id"`
	Name   string `json:"name" bson:"name"`
	*User         // 指针类型的嵌入字段
	*Order        // 指针类型的嵌入字段
}

func main() {
	fmt.Println("=== 测试 1: 嵌入对象的值类型 ===")
	testEmbeddedValue()

	fmt.Println("\n=== 测试 2: 嵌入对象的指针类型 ===")
	testEmbeddedPointer()

	fmt.Println("\n=== 测试 3: 嵌套的嵌入对象 ===")
	testNestedEmbedded()

	fmt.Println("\n=== 测试 4: 临时结构体 ===")
	testTempStruct()
}

// 测试嵌入对象的值类型
func testEmbeddedValue() {
	user := User{
		Name: "John",
		Age:  30,
		Address: Address{
			City:    "New York",
			Street:  "123 Main St",
			ZipCode: "10001",
		},
	}

	sch, err := schema.GetOrParse(&user, nil)
	if err != nil {
		fmt.Printf("Error parsing schema: %v\n", err)
		return
	}

	fmt.Println("获取值测试:")
	fmt.Printf("Name: %v\n", sch.GetValue(&user, "Name"))
	fmt.Printf("Age: %v\n", sch.GetValue(&user, "Age"))
	fmt.Printf("City: %v\n", sch.GetValue(&user, "Address", "City"))
	fmt.Printf("Street: %v\n", sch.GetValue(&user, "Address", "Street"))

	fmt.Println("\n设置值测试:")
	_ = sch.SetValue(&user, "Jane", "Name")
	_ = sch.SetValue(&user, 31, "Age")
	_ = sch.SetValue(&user, "Los Angeles", "Address", "City")
	_ = sch.SetValue(&user, "456 Oak Ave", "Address", "Street")

	fmt.Printf("Updated Name: %v\n", user.Name)
	fmt.Printf("Updated Age: %v\n", user.Age)
	fmt.Printf("Updated City: %v\n", user.Address.City)
	fmt.Printf("Updated Street: %v\n", user.Address.Street)
}

// 测试嵌入对象的指针类型
func testEmbeddedPointer() {
	address := Address{
		City:    "Boston",
		Street:  "789 Elm St",
		ZipCode: "02101",
	}
	user := &User{Name: "Alice", Age: 25, Address: address}
	order := &Order{ID: 1001, UserID: 1, Amount: 99.99}
	customer := Customer{ID: 1, Name: "Alice Smith", User: user, Order: order}

	sch, err := schema.GetOrParse(&customer, nil)
	if err != nil {
		fmt.Printf("Error parsing schema: %v\n", err)
		return
	}

	fmt.Println("获取值测试:")
	fmt.Printf("Customer Name: %v\n", sch.GetValue(&customer, "Name"))
	fmt.Printf("User Name: %v\n", sch.GetValue(&customer, "User", "Name"))
	fmt.Printf("User Age: %v\n", sch.GetValue(&customer, "User", "Age"))
	fmt.Printf("Order ID: %v\n", sch.GetValue(&customer, "Order", "ID"))
	fmt.Printf("Order Amount: %v\n", sch.GetValue(&customer, "Order", "Amount"))

	fmt.Println("\n设置值测试:")
	_ = sch.SetValue(&customer, "Alice Johnson", "Name")
	_ = sch.SetValue(&customer, "Alicia", "User", "Name")
	_ = sch.SetValue(&customer, 26, "User", "Age")
	_ = sch.SetValue(&customer, 1002, "Order", "ID")
	_ = sch.SetValue(&customer, 199.99, "Order", "Amount")

	fmt.Printf("Updated Customer Name: %v\n", customer.Name)
	fmt.Printf("Updated User Name: %v\n", customer.User.Name)
	fmt.Printf("Updated User Age: %v\n", customer.User.Age)
	fmt.Printf("Updated Order ID: %v\n", customer.Order.ID)
	fmt.Printf("Updated Order Amount: %v\n", customer.Order.Amount)
}

// 测试嵌套的嵌入对象
func testNestedEmbedded() {
	nestedStruct := struct {
		User
		Order
		Notes string `json:"notes" bson:"notes"`
	}{
		User: User{
			Name: "Bob", Age: 35,
			Address: Address{City: "Chicago", Street: "321 Pine St", ZipCode: "60601"},
		},
		Order: Order{ID: 2001, UserID: 2, Amount: 149.99},
		Notes: "Test order",
	}

	sch, err := schema.GetOrParse(&nestedStruct, nil)
	if err != nil {
		fmt.Printf("Error parsing schema: %v\n", err)
		return
	}

	fmt.Println("获取值测试:")
	fmt.Printf("Name: %v\n", sch.GetValue(&nestedStruct, "Name"))
	fmt.Printf("Age: %v\n", sch.GetValue(&nestedStruct, "Age"))
	fmt.Printf("City: %v\n", sch.GetValue(&nestedStruct, "Address", "City"))
	fmt.Printf("Order ID: %v\n", sch.GetValue(&nestedStruct, "ID"))
	fmt.Printf("Amount: %v\n", sch.GetValue(&nestedStruct, "Amount"))
	fmt.Printf("Notes: %v\n", sch.GetValue(&nestedStruct, "Notes"))

	fmt.Println("\n设置值测试:")
	_ = sch.SetValue(&nestedStruct, "Robert", "Name")
	_ = sch.SetValue(&nestedStruct, 36, "Age")
	_ = sch.SetValue(&nestedStruct, "Houston", "Address", "City")
	_ = sch.SetValue(&nestedStruct, 2002, "ID")
	_ = sch.SetValue(&nestedStruct, 199.99, "Amount")
	_ = sch.SetValue(&nestedStruct, "Updated test order", "Notes")

	fmt.Printf("Updated Name: %v\n", nestedStruct.Name)
	fmt.Printf("Updated Age: %v\n", nestedStruct.Age)
	fmt.Printf("Updated City: %v\n", nestedStruct.Address.City)
	fmt.Printf("Updated Order ID: %v\n", nestedStruct.ID)
	fmt.Printf("Updated Amount: %v\n", nestedStruct.Amount)
	fmt.Printf("Updated Notes: %v\n", nestedStruct.Notes)
}

// 测试临时结构体
func testTempStruct() {
	args := struct {
		User
		Order
		Iid  int32    `json:"Iid" bson:"Iid"`
		Tags []string `json:"tags" bson:"tags"`
	}{
		User: User{
			Name: "Charlie", Age: 40,
			Address: Address{City: "Seattle", Street: "987 Cedar St", ZipCode: "98101"},
		},
		Order: Order{ID: 3001, UserID: 3, Amount: 249.99},
		Iid:   12345,
		Tags:  []string{"test", "temp"},
	}

	sch, err := schema.GetOrParse(&args, nil)
	if err != nil {
		fmt.Printf("Error parsing schema: %v\n", err)
		return
	}

	fmt.Println("获取值测试:")
	fmt.Printf("Name: %v\n", sch.GetValue(&args, "Name"))
	fmt.Printf("Age: %v\n", sch.GetValue(&args, "Age"))
	fmt.Printf("City: %v\n", sch.GetValue(&args, "Address", "City"))
	fmt.Printf("Order ID: %v\n", sch.GetValue(&args, "ID"))
	fmt.Printf("Amount: %v\n", sch.GetValue(&args, "Amount"))
	fmt.Printf("Iid: %v\n", sch.GetValue(&args, "Iid"))
	fmt.Printf("Tags: %v\n", sch.GetValue(&args, "Tags"))

	fmt.Println("\n设置值测试:")
	_ = sch.SetValue(&args, "Charles", "Name")
	_ = sch.SetValue(&args, 41, "Age")
	_ = sch.SetValue(&args, "Portland", "Address", "City")
	_ = sch.SetValue(&args, 3002, "ID")
	_ = sch.SetValue(&args, 299.99, "Amount")
	_ = sch.SetValue(&args, int32(67890), "Iid")
	_ = sch.SetValue(&args, []string{"test", "temp", "updated"}, "Tags")

	fmt.Printf("Updated Name: %v\n", args.Name)
	fmt.Printf("Updated Age: %v\n", args.Age)
	fmt.Printf("Updated City: %v\n", args.Address.City)
	fmt.Printf("Updated Order ID: %v\n", args.ID)
	fmt.Printf("Updated Amount: %v\n", args.Amount)
	fmt.Printf("Updated Iid: %v\n", args.Iid)
	fmt.Printf("Updated Tags: %v\n", args.Tags)
}
