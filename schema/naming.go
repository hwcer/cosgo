package schema

import (
	"strings"
)

// Namer Namer interface
type Namer interface {
	TableName(table string) string
	ColumnName(table, column string) string
}

// Replacer replacer interface like strings.Replacer
type Replacer interface {
	Replace(name string) string
}

// NamingStrategy tables, columns naming strategy
type NamingStrategy struct {
	TablePrefix   string
	SingularTable bool
	NameReplacer  Replacer
	NoLowerCase   bool
}

// TableName convert string to table name
func (ns NamingStrategy) TableName(str string) string {
	return strings.ToLower(str)
}

// ColumnName convert string to column name
func (ns NamingStrategy) ColumnName(table, column string) string {
	return ns.toDBName(column)
}

var (
	// https://github.com/golang/lint/blob/master/lint.go#L770
	commonInitialisms         = []string{"API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "LHS", "QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SSH", "TLS", "TTL", "UID", "UI", "UUID", "URI", "URL", "UTF8", "VM", "XML", "XSRF", "XSS"}
	commonInitialismsReplacer *strings.Replacer
)

func init() {
	commonInitialismsForReplacer := make([]string, 0, len(commonInitialisms))
	for _, initialism := range commonInitialisms {
		lower := strings.ToLower(initialism)
		titled := strings.ToUpper(lower[:1]) + lower[1:]
		commonInitialismsForReplacer = append(commonInitialismsForReplacer, initialism, titled)
	}
	commonInitialismsReplacer = strings.NewReplacer(commonInitialismsForReplacer...)
}

func (ns NamingStrategy) toDBName(name string) string {
	if name == "" {
		return ""
	}

	if ns.NameReplacer != nil {
		name = ns.NameReplacer.Replace(name)
	}

	if ns.NoLowerCase {
		return name
	}

	var (
		value                          = commonInitialismsReplacer.Replace(name)
		buf                            strings.Builder
		lastCase, nextCase, nextNumber bool // upper case == true
		curCase                        = value[0] <= 'Z' && value[0] >= 'A'
	)

	for i, v := range value[:len(value)-1] {
		nextCase = value[i+1] <= 'Z' && value[i+1] >= 'A'
		nextNumber = value[i+1] >= '0' && value[i+1] <= '9'

		if curCase {
			if lastCase && (nextCase || nextNumber) {
				buf.WriteRune(v + 32)
			} else {
				if i > 0 && value[i-1] != '_' && value[i+1] != '_' {
					buf.WriteByte('_')
				}
				buf.WriteRune(v + 32)
			}
		} else {
			buf.WriteRune(v)
		}

		lastCase = curCase
		curCase = nextCase
	}

	if curCase {
		if !lastCase && len(value) > 1 {
			buf.WriteByte('_')
		}
		buf.WriteByte(value[len(value)-1] + 32)
	} else {
		buf.WriteByte(value[len(value)-1])
	}
	return buf.String()
}
