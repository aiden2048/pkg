package mongodb

import "go.mongodb.org/mongo-driver/bson"

type Filter bson.D

func (m *Filter) Add(key string, value any) {
	*m = append(*m, bson.E{Key: key, Value: value})
}
func (m Filter) M() bson.M {
	bm := make(bson.M, len(m))
	for _, e := range m {
		bm[e.Key] = e.Value
	}
	return bm
}
func (m Filter) D() Filter {
	return m
}
