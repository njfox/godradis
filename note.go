package godradis

import "github.com/iancoleman/orderedmap"

type Note struct {
	Id int `json:"id"`
	CategoryId int `json:"category_id"`
	Title string `json:"title"`
	Fields orderedmap.OrderedMap `json:"fields"`
	Text string `json:"text"`
	Node *Node
}