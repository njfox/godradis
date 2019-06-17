package godradis

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