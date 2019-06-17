package godradis

type Team struct {
	Id int `json:"id"`
	Name string `json:"name"`
	TeamSince string `json:"client_since"` // TODO: update this if the API gets fixed (it should return "team_since")
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Projects []TeamProject `json:"projects"`
}

type TeamProject struct {
	Id int `json:"id"`
	Name string `json:"name"`
}