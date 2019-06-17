package godradis

type Client struct {
	Id int `json:"id"`
	Name string `json:"name"`
}

type Author struct {
	Email string `json:"email"`
}

type Owner struct {
	Email string `json:"email"`
}

type Project struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Client Client `json:"client"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Authors []Author `json:"authors"`
	Owners []Owner `json:"owners"`
}