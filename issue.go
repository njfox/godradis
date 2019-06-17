package godradis

import "github.com/iancoleman/orderedmap"

type Issue struct {
	Id int `json:"id"`
	Title string `json:"title"`
	Fields orderedmap.OrderedMap `json:"fields"`
	Text string `json:"text"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Project *Project
}
