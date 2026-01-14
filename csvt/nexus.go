package csvt

import "github.com/Rafael24595/go-collections/collection"

type nexus struct {
	key   string
	root  bool
	nodes collection.Vector[group]
}

func newNexus(key string, root bool, nodes []group) nexus {
	return nexus{
		key:   key,
		root:  root,
		nodes: *collection.VectorFromList(nodes),
	}
}

func (r *nexus) get(position int) (*group, bool) {
	group, ok := r.nodes.Get(position)
	return &group, ok
}
