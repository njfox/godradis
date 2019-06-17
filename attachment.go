package godradis

type Attachment struct {
	Filename string `json:"filename"`
	Link string `json:"link"`
	Node *Node
}