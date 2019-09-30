package godradis

import "github.com/iancoleman/orderedmap"

type IssueLib struct {
	Id int `json:"id"`
	Title string `json:"title"`
	Fields orderedmap.OrderedMap `json:"fields"`
	State int `json:"state"`
	Content string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}