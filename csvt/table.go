package csvt

import "github.com/Rafael24595/go-collections/collection"

type table struct {
	nexus collection.Dictionary[string, nexus]
}

func newTable(nexus map[string]nexus) table {
	tbl := collection.DictionaryFromMap(nexus)
	return table{
		nexus: *tbl,
	}
}

func (r *table) root() (*nexus, bool) {
	nexus, ok := r.nexus.FindOne(func(k string, rn nexus) bool {
		return rn.root
	})
	return &nexus, ok
}

func (r *table) Find(node *node) (*group, bool) {
	value, exists := r.nexus.Get(node.key())
	if !exists {
		return nil, false
	}
	if node.index != -1 {
		group, ok := value.nodes.Get(node.index)
		return &group, ok
	}
	return nil, false
}
