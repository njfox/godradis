package godradis

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

type Node struct {
	Id int `json:"id"`
	Label string `json:"label"`
	TypeId int `json:"type_id"`
	ParentId int `json:"parent_id"`
	Position int `json:"position"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Evidence []Evidence `json:"evidence"`
	Notes []Note `json:"notes"`
	Project *Project
}

func (n *Node) GetEvidenceById(id int) (*Evidence, error) {
	for i, evidence := range n.Evidence {
		if evidence.Id == id {
			return &n.Evidence[i], nil
		}
	}
	return nil, errors.New(fmt.Sprintf("no evidence on node for id %v", id))
}

func (n *Node) GetEvidenceByIssueTitle(title string) []*Evidence {
	var evidenceInstances []*Evidence
	for i, evidence := range n.Evidence {
		if strings.ToLower(evidence.Issue.Title) == strings.ToLower(title) {
			evidenceInstances = append(evidenceInstances, &n.Evidence[i])
		}
	}
	return evidenceInstances
}

func (n *Node) GetEvidenceByField(key, value string) []*Evidence {
	var evidenceInstances []*Evidence
	for i, evidence := range n.Evidence {
		val, ok := evidence.Fields.Get(key)
		if !ok {
			continue
		}
		if val == value {
			evidenceInstances = append(evidenceInstances, &n.Evidence[i])
		}
	}
	return evidenceInstances
}

func (n *Node) GetNoteById(id int) (*Note, error) {
	for i, note := range n.Notes {
		if note.Id == id {
			return &n.Notes[i], nil
		}
	}
	return nil, errors.New(fmt.Sprintf("no note on node for id %v", id))
}

func (n *Node) GetNotesByTitle(title string) []*Note {
	var notes []*Note
	for i, note := range n.Notes {
		if strings.ToLower(note.Title) == strings.ToLower(title) {
			notes = append(notes, &n.Notes[i])
		}
	}
	return notes
}

func (n *Node) setEvidenceNodeReferences() {
	for i := range n.Evidence {
		n.setEvidenceNodeReference(&n.Evidence[i])
	}
}

func (n *Node) setEvidenceNodeReference(e *Evidence) {
	e.Node = n
}

func (n *Node) addEvidence(e Evidence) {
	n.Evidence = append(n.Evidence, e)
}

func (n *Node) deleteEvidence(e Evidence) {
	for i, evidence := range n.Evidence {
		if evidence.Id == e.Id {
			copy(n.Evidence[i:], n.Evidence[i+1:])
		}
	}
}

func (n *Node) setNoteNodeReferences() {
	for i := range n.Notes {
		n.setNoteNodeReference(&n.Notes[i])
	}
}

func (n *Node) setNoteNodeReference(note *Note) {
	note.Node = n
}

func (n *Node) addNote(note Note) {
	n.Notes = append(n.Notes, note)
}

func (n *Node) deleteNote(note Note) {
	for i, _note := range n.Notes {
		if _note.Id == note.Id {
			copy(n.Notes[i:], n.Notes[i+1:])
		}
	}
}