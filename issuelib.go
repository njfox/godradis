package godradis

import (
	"fmt"
	"github.com/iancoleman/orderedmap"
	"github.com/pkg/errors"
)

type IssueLib struct {
	Id int `json:"id"`
	Title string `json:"title"`
	Fields orderedmap.OrderedMap `json:"fields"`
	State int `json:"state"`
	Content string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (i *IssueLib) SetField(key, value string) {
	i.Fields.Set(key, value)
}

func (i *IssueLib) GetField(key string) (string, error) {
	value, ok := i.Fields.Get(key)
	if !ok {
		return "", errors.New(fmt.Sprintf("field not found: %v", key))
	}
	return value.(string), nil
}

func (i *IssueLib) CopyFields() orderedmap.OrderedMap {
	fields := orderedmap.New()
	keys := i.Fields.Keys()
	for _, k := range keys {
		value, ok := i.Fields.Get(k)
		if ok {
			fields.Set(k, value)
		}
	}
	return *fields
}