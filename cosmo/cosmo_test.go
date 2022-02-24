package cosmo

import (
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

type Role struct {
	Id    string `bson:"_id" gorm:"column:_id;primaryKey"`
	Name  string `gorm:"index:idx_name,sort:desc;index:,unique"`
	Login int64  `gorm:"index:idx_name,sort:asc;index:,sparse"`
}

func TestCosmo(t *testing.T) {
	db := New("hwc")
	var err error
	if err = db.Start("10.26.17.20:27017"); err != nil {
		t.Logf("%v", err)
		return
	}

	//if err := db.AutoMigrator(&Role{}); err != nil {
	//	t.Logf("AutoMigrator Error:%v", err)
	//}

	t.Logf("================Find=====================")
	var roles []*Role
	db = db.Table("role").Omit("login").Page(1, 2).Order("_id", -1).Find(&roles)
	if db.Error != nil {
		t.Logf("Find error:%v", db.Error)
	} else {
		t.Logf("RowsAffected:%v", db.RowsAffected)
		for _, v := range roles {
			t.Logf("role:%+v", v)
		}
	}
	t.Logf("==================Update1===================")
	role := &Role{
		Id: "2d71",
	}
	update := bson.M{"Name": "222"}
	update["$inc"] = bson.M{"login": 1}

	db = db.Model(role).Update(update)
	if db.Error != nil {
		t.Logf("%v", db.Error)
	} else {
		t.Logf("RowsAffected:%v,role:%+v", db.RowsAffected, role)
	}

	t.Logf("=================Update2====================")
	role2 := &Role{
		Id:    "2d71",
		Login: 3,
		Name:  "sssss",
	}
	db = db.Table("role").Select("name").Update(role2)
	if db.Error != nil {
		t.Logf("%v", db.Error)
	} else {
		t.Logf("RowsAffected:%v", db.RowsAffected)
	}

}
