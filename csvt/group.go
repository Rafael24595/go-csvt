package csvt

import (
	"strconv"

	"github.com/Rafael24595/go-collections/collection"
)

type group struct {
	category category
	headers  collection.Vector[string]
	group    any
}

func newGroup[T any](category category, headers []string, grp T) group {
	return group{
		category: category,
		headers:  *collection.VectorFromList(headers),
		group:    grp,
	}
}

func (r *group) findField(key string) (*node, bool) {
	switch v := r.group.(type) {
	case []node:
		index := r.headers.IndexOf(func(s string) bool {
			return s == key
		})
		if index == -1 || index > len(v) {
			return nil, false
		}
		return &v[index], true
	default:
		return nil, false
	}
}

func (r *group) findFields() []collection.Pair[string, node] {
	pairs := []collection.Pair[string, node]{}
	switch v := r.group.(type) {
	case map[string]node:
		for k, v := range v {
			pairs = append(pairs, collection.NewPair(k, v))
		}
	case []node:
		for i, v := range v {
			pairs = append(pairs, collection.NewPair(strconv.Itoa(i), v))
		}
	}
	return pairs
}

func (r *group) findValue() (*node, bool) {
	switch v := r.group.(type) {
	case node:
		return &v, true
	default:
		return nil, false
	}
}
