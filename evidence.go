package godradis

import "github.com/iancoleman/orderedmap"

type Evidence struct {
	Id int `json:"id"`
	Content string `json:"content"`
	Fields orderedmap.OrderedMap `json:"fields"`
	Issue EvidenceIssue `json:"issue"`
	Node *Node
}

type EvidenceIssue struct {
	Id int `json:"id"`
	Title string `json:"title"`
	Url string `json:"url"`
}