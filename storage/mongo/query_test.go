package mongo

import "testing"

func TestNewQuery(t *testing.T) {
	query := make(Query)
	query.Eq("a", 1)
	query.Gt("a", 1)
	query.Lte("lte", 1)
	t.Logf("query:%v", query)

	query1 := NewPrimaryQuery("a", 1)
	query2 := NewPrimaryQuery("b", 2)

	query3 := make(Query)
	query3.AND(query, query1, query2)
	t.Logf("query3:%v", query3)
}
