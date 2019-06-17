package godradis

import (
	"fmt"
	"github.com/iancoleman/orderedmap"
	"github.com/pkg/errors"
)

type Note struct {
	Id int `json:"id"`
	CategoryId int `json:"category_id"`
	Title string `json:"title"`
	Fields orderedmap.OrderedMap `json:"fields"`
	Text string `json:"text"`
	Node *Node
}

func (n *Note) SetField(key, value string) {
	n.Fields.Set(key, value)
}

func (n *Note) GetField(key string) (string, error) {
	value, ok := n.Fields.Get(key)
	if !ok {
		return "", errors.New(fmt.Sprintf("field not found: %v", key))
	}
	return value.(string), nil
}