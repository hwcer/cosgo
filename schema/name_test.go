package schema

import "testing"

// dbnameLeaf 叶子结构:字段名 PascalCase、无 bson 标签,DBName 回落成小写。
// 混入带 bson / json 标签的字段,验证标签优先于回落,且两套标签互不串味。
type dbnameLeaf struct {
	Lv    int32
	Exp   int64  `bson:"experience" json:"totalExp"`
	Owner string `json:"OwnerName"`
	Tail  int32  `form:"tail_form"`
}

type DbnameEmbed struct {
	Extra int32
}

// dbnameRoot 覆盖需要逐段判定的四种容器:map / slice / 嵌套结构体 / 标量。
type dbnameRoot struct {
	DbnameEmbed
	BreakLv    int32
	SoulRelics map[int32]*dbnameLeaf
	Tags       map[string]string
	Slots      []*dbnameLeaf
	Profile    *dbnameLeaf
	Counters   map[int32]int64
}

func TestSchemaDBName(t *testing.T) {
	sch, err := Parse(&dbnameRoot{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	cases := map[string]string{
		//整字段:大小写都归一到落库名
		"BreakLv": "breaklv",
		"breaklv": "breaklv",

		//map 键原样保留,再往下按值类型继续换名
		"SoulRelics":      "soulrelics",
		"SoulRelics.1":    "soulrelics.1",
		"SoulRelics.1.Lv": "soulrelics.1.lv",

		//🔴 map 键的大小写有业务含义,不得被动
		"Tags.AbC": "tags.AbC",

		//slice 下标同理
		"Slots.0.Lv": "slots.0.lv",

		//嵌套结构体逐段换名
		"Profile.Lv": "profile.lv",

		//bson 标签优先于小写回落
		"Profile.Exp": "profile.experience",

		//🔴 无 bson 标签时不看 json 标签,仍按字段名小写(Owner 的 json 标签是 OwnerName)
		"Profile.Owner": "profile.owner",

		//匿名嵌入字段在顶层被提升
		"Extra": "extra",
	}
	for path, want := range cases {
		got, err := sch.DBName(path)
		if err != nil {
			t.Errorf("DBName(%q) 不该报错: %v", path, err)
			continue
		}
		if got != want {
			t.Errorf("DBName(%q) = %q, 期望 %q", path, got, want)
		}
	}
}

// JSName 与 DBName 共用同一套逐段走查，差别只在每段取哪个名字。
func TestSchemaJSName(t *testing.T) {
	sch, err := Parse(&dbnameRoot{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	cases := map[string]string{
		//无 json 标签 → 回落成 Go 字段名(不小写，这点与 DBName 不同)
		"BreakLv":    "BreakLv",
		"breaklv":    "BreakLv",
		"Profile.Lv": "Profile.Lv",

		//map 键原样保留，再往下继续按 json 标签换名
		"SoulRelics.1":     "SoulRelics.1",
		"SoulRelics.1.Lv":  "SoulRelics.1.Lv",
		"SoulRelics.1.Exp": "SoulRelics.1.totalExp",

		//🔴 json 标签优先，且不受 bson 标签影响(Exp 的 bson 标签是 experience)
		"Profile.Exp":   "Profile.totalExp",
		"Profile.Owner": "Profile.OwnerName",

		//map 键大小写照样不动
		"Tags.AbC": "Tags.AbC",
	}
	for path, want := range cases {
		got, err := sch.JSName(path)
		if err != nil {
			t.Errorf("JSName(%q) 不该报错: %v", path, err)
			continue
		}
		if got != want {
			t.Errorf("JSName(%q) = %q, 期望 %q", path, got, want)
		}
	}
}

// 自定义标签走 Field.GetName：取到标签用标签，取不到回落成 Go 字段名（不小写）。
func TestSchemaNameCustomTag(t *testing.T) {
	sch, err := Parse(&dbnameRoot{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	cases := map[string]string{
		"Profile.Tail": "Profile.tail_form", //命中 form 标签
		"Profile.Lv":   "Profile.Lv",        //无 form 标签 → Go 字段名
	}
	for path, want := range cases {
		got, err := sch.GetName(path, "form")
		if err != nil {
			t.Errorf("GetName(%q,form) 不该报错: %v", path, err)
			continue
		}
		if got != want {
			t.Errorf("GetName(%q,form) = %q, 期望 %q", path, got, want)
		}
	}
}

// 路径写错必须报错，而不是原样透传。透传等于让 mongo 的 $set 往库里插一个野字段，
// 真正的字段纹丝不动，且没有任何报错 —— 静默丢数据。
func TestSchemaDBNameRejectsBadPath(t *testing.T) {
	sch, err := Parse(&dbnameRoot{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bad := map[string]string{
		"nosuchfield":       "根字段不存在",
		"nosuchfield.1":     "根字段不存在(多级)",
		"SoulRelics.1.Nope": "map 值结构体里没有这个字段",
		"Profile.Nope":      "嵌套结构体里没有这个字段",
		"Counters.1.Lv":     "map 值是标量,不能再往下钻",
		"BreakLv.1":         "标量字段不能再往下钻",
		"SoulRelics.":       "空段",
		"":                  "空路径",
	}
	for path, why := range bad {
		if got, err := sch.DBName(path); err == nil {
			t.Errorf("DBName(%q) 应报错(%s)，实际返回 %q", path, why, got)
		}
		//JSName 走同一套走查，越界判定必须一致
		if got, err := sch.JSName(path); err == nil {
			t.Errorf("JSName(%q) 应报错(%s)，实际返回 %q", path, why, got)
		}
	}
}
