package godradis

import (
	"errors"
	"fmt"
	"github.com/iancoleman/orderedmap"
)

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

func (e *Evidence) SetField(key, value string) {
	e.Fields.Set(key, value)
}

func (e *Evidence) GetField(key string) (string, error) {
	value, ok := e.Fields.Get(key)
	if !ok {
		return "", errors.New(fmt.Sprintf("field not found: %v", key))
	}
	return value.(string), nil
}

func (e *Evidence) CopyFields() orderedmap.OrderedMap {
	fields := orderedmap.New()
	keys := e.Fields.Keys()
	for _, k := range keys {
		value, ok := e.Fields.Get(k)
		if ok {
			fields.Set(k, value)
		}
	}
	return *fields
}