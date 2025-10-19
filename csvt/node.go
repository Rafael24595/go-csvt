package csvt

import "fmt"

type node struct {
	value interface{}
	index int
}

func fromPointer(value interface{}, index int) node {
	return node{
		value: value,
		index: index,
	}
}

func fromNonPointer(value interface{}) node {
	return node{
		value: value,
		index: -1,
	}
}

func fromEmpty() node {
	return fromNonPointer("")
}

func (n node) key() string {
	return fmt.Sprintf("%v", n.value)
}
