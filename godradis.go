// Package godradis provides a full-featured library for accessing the Dradis server REST API.
package godradis

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/iancoleman/orderedmap"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Godradis struct {
	Config Config
	httpClient http.Client
}

// Configuration

type Config struct {
	BaseUrl string `json:"dradis_url"`
	ApiKey string `json:"api_key"`
	Verify bool `json:"verify"`
}

/*
Configure populates godradis with parameters necessary to access the Dradis server. url should be the base url of the
Dradis server, before "/pro/api" and without a trailing "/" (e.g. https://example.com). apiKey is a string containing
the API key shown on the user's Dradis profile page. verify instructs godradis whether to check TLS certificates on the
Dradis server.

After creating the configuration, Configure creates an http.client on the Godradis object to be used for all subsequent
HTTP requests to the Dradis server.

    gd := godradis.Godradis{}
    gd.Configure("https://example.com", "abcdefghijk", false)
 */
func (gd *Godradis) Configure(url, apiKey string, verify bool) {
	gd.Config = Config{url, apiKey, verify}
	gd.createClient(verify)
}

/*
LoadConfig behaves the same way as Configure except that it loads the configuration parameters from a JSON file instead
of accepting them directly in the function call.

    gd := godradis.Godradis{}
    err := gd.LoadConfig("dradis_config.json")
    if err != nil {
        fmt.Println(err)
    }
 */
func (gd *Godradis) LoadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	fileBytes, _ := ioutil.ReadAll(file)
	err = json.Unmarshal(fileBytes, &gd.Config)
	if err != nil {
		return err
	}
	gd.createClient(gd.Config.Verify)
	return nil
}

// Utils

func (gd *Godradis) createClient(verify bool) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !verify},
	}

	gd.httpClient = http.Client{Transport: tr}
}

func (gd *Godradis) sendRequest(method, resource string, body []byte) (*http.Response, error) {
	req, _ := http.NewRequest(method, fmt.Sprintf("%s/pro/api/%s", gd.Config.BaseUrl, resource), bytes.NewBuffer(body))
	req.Header.Add("Authorization", fmt.Sprintf(`Token token="%s"`, gd.Config.ApiKey))
	if method == "DELETE" || ((method == "POST" || method == "PUT") && body != nil) {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := gd.httpClient.Do(req)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (gd *Godradis) sendRequestWithProjectId(method, resource string, projectId int, body []byte) (*http.Response, error) {
	req, _ := http.NewRequest(method, fmt.Sprintf("%s/pro/api/%s", gd.Config.BaseUrl, resource), bytes.NewBuffer(body))
	req.Header.Add("Authorization", fmt.Sprintf(`Token token="%s"`, gd.Config.ApiKey))
	if method == "DELETE" || ((method == "POST" || method == "PUT") && body != nil) {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Dradis-Project-Id", strconv.Itoa(projectId))
	resp, err := gd.httpClient.Do(req)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func parseOrderedMapFields(fields *orderedmap.OrderedMap) string {
	text := ""
	keys := fields.Keys()
	for _, k := range keys {
		v, _ := fields.Get(k)
		text += fmt.Sprintf("#[%v]#\r\n%v\r\n\r\n", k, v)
	}
	return text
}

// Projects Endpoint

/*
GetAllProjects takes no arguments and retrieves a full list of all projects on the Dradis server. If an error of any kind
occurs, the function will return an empty rather than partial list as well as the error.

    gd := godradis.Godradis{}

    [...]

    projectList, _ := gd.GetAllProjects()
    if len(projectList) > 0 {
        fmt.Printf("%v", projectList[0].Name)
    }
 */
func (gd *Godradis) GetAllProjects() ([]Project, error) {
	resp, err := gd.sendRequest("GET", "projects", nil)
	if err != nil {
		return []Project{}, err
	}
	defer resp.Body.Close()
	var projects []Project
	if resp.StatusCode != http.StatusOK {
		return []Project{}, errors.New("could not get projects from server")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []Project{}, err
	}

	err = json.Unmarshal(body, &projects)
	if err != nil {
		return []Project{}, err
	}
	return projects, nil
}

/*
GetProjectById fetches a Project object from the Dradis server based on the int id.

    gd := godradis.Godradis{}

    [...]

    project, err := gd.GetProjectById(45)
    if err != nil {
        fmt.Println(err)
    }
 */
func (gd *Godradis) GetProjectById(id int) (Project, error) {
	resp, err := gd.sendRequest("GET", fmt.Sprintf("projects/%v", id), nil)
	if err != nil {
		return Project{}, err
	}
	defer resp.Body.Close()
	var project Project
	if resp.StatusCode != http.StatusOK {
		return Project{}, errors.New("could not get project from server")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Project{}, err
	}

	err = json.Unmarshal(body, &project)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

/*
GetProjectByName searches for and returns a Project object based on the name. GetProjectByName works by calling GetAllProjects
first and then ranges over them comparing the name strings.

    gd := godradis.Godradis{}

    [...]

    project, err := gd.GetProjectByName("Foobar External Network Penetration Test")
    if err != nil {
        fmt.Println(err)
    }
 */
func (gd *Godradis) GetProjectByName(name string) (Project, error) {
	projects, err := gd.GetAllProjects()
	if err != nil {
		return Project{}, err
	}
	for _, project := range projects {
		if strings.ToLower(project.Name) == strings.ToLower(name) {
			return project, nil
		}
	}
	return Project{}, errors.New(fmt.Sprintf("Could not find project %s", name))
}

type projectDetails struct {
	Name string `json:"name,omitempty"`
	ClientId int `json:"team_id,omitempty"` // For some reason, POST/PUT methods use strings instead of ints even though they return ints
	ReportTemplatePropertiesId int `json:"report_template_properties_id,omitempty"`
	AuthorIds []int `json:"author_ids,omitempty"`
	Template string `json:"template,omitempty"`
}

func (pd *projectDetails) parseArguments(name, clientId, reportTemplatePropertiesId interface{}, authorIds []int, template interface{}) {
	if name == nil {
		pd.Name = ""
	} else {
		pd.Name = name.(string)
	}
	if clientId == nil {
		pd.ClientId = 0
	} else {
		pd.ClientId = clientId.(int)
	}
	if reportTemplatePropertiesId == nil {
		pd.ReportTemplatePropertiesId = 0
	} else {
		pd.ReportTemplatePropertiesId = reportTemplatePropertiesId.(int)
	}
	pd.AuthorIds = authorIds
	if template == nil {
		pd.Template = ""
	} else {
		pd.Template = template.(string)
	}
}

/*
CreateProject creates a project on the Dradis server and returns the newly created Project object. All 5 arguments are
required in the function call, but only name and clientId must be non-nil. reportTemplatePropertiesId is an optional int
that assigns a default report template to the project. authorIds accepts an int slice of authors to assign to the project.
template is an optional string that assigns the project template based on the template name.

    gd := godradis.Godradis{}

    [...]

    authors := [2]int[]{3, 4}
    project, _ := gd.CreateProject("New Project Name", 1, nil, authors[:], nil)
    fmt.Printf("%v", project.Name)
 */
func (gd *Godradis) CreateProject(name string, clientId int, reportTemplatePropertiesId interface{}, authorIds []int, template interface{}) (Project, error) {
	// Required so that json.Marshal() sends the project fields wrapped in a project{} json object
	type reqModel struct {
		Pd projectDetails `json:"project"`
	}

	pd := projectDetails{}
	pd.parseArguments(name, clientId, reportTemplatePropertiesId, authorIds, template)

	jsonBody, err := json.Marshal(&reqModel{pd})
	if err != nil {
		return Project{}, err
	}
	resp, err := gd.sendRequest("POST", "projects", jsonBody)
	if err != nil {
		return Project{}, err
	}
	defer resp.Body.Close()
	var newProject Project
	if resp.StatusCode != http.StatusCreated {
		return Project{}, errors.New("could not create project")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Project{}, err
	}

	err = json.Unmarshal(body, &newProject)
	if err != nil {
		return Project{}, err
	}

	return newProject, nil
}

/*
UpdateProject takes a reference to an existing Project object as well as 5 arguments representing properties to update.
All arguments are required to be passed to UpdateProject but only properties being modified need to be non-nil. UpdateProject
modifies the original Project object in-place rather than returning a new one.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.CreateProject("New Project Name", 1, nil, nil, nil)
    err := gd.UpdateProject(&project, "Modified the project name", nil, nil, nil, nil)
    if err != nil {
        fmt.Printf("%v", project.Name)
    }
 */
func (gd *Godradis) UpdateProject(p *Project, name, clientId, reportTemplatePropertiesId interface{}, authorIds []int, template interface{}) error {
	// Required so that json.Marshal() sends the project fields wrapped in a project{} json object
	type reqModel struct {
		Pd projectDetails `json:"project"`
	}

	pd := projectDetails{}
	pd.parseArguments(name, clientId, reportTemplatePropertiesId, authorIds, template)

	jsonBody, err := json.Marshal(&reqModel{pd})
	if err != nil {
		return err
	}
	resp, err := gd.sendRequest("PUT", fmt.Sprintf("projects/%v", p.Id), jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("could not update project")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &p)
	if err != nil {
		return err
	}
	return nil
}

/*
DeleteProject takes a reference to a Project object and deletes the project on the Dradis server.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.CreateProject("New Project Name", 1, nil, nil, nil)
    err := gd.DeleteProject(&project)
    if err != nil {
        fmt.Println(err)
    }
 */
func (gd *Godradis) DeleteProject(p *Project) error {
	resp, err := gd.sendRequest("DELETE", fmt.Sprintf("projects/%v", p.Id), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return errors.New("could not delete project.")
	}
}

// Teams endpoint

/*
GetAllTeams takes no arguments and returns a list of all teams on the server.

    gd := godradis.Godradis{}

    [...]

    teamList, _ := gd.GetAllTeams()
    if len(teamList) > 0 {
        fmt.Printf("%v", teamList[0].Name)
    }
 */
func (gd *Godradis) GetAllTeams() ([]Team, error) {
	resp, err := gd.sendRequest("GET", "teams", nil)
	if err != nil {
		return []Team{}, err
	}
	defer resp.Body.Close()
	var teams []Team
	if resp.StatusCode != http.StatusOK {
		return []Team{}, errors.New("could not get teams list")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []Team{}, err
	}

	err = json.Unmarshal(body, &teams)
	if err != nil {
		return []Team{}, err
	}
	return teams, nil
}

/*
GetTeamById fetches a Team object from the Dradis server based on the int id.

    gd := godradis.Godradis{}

    [...]

    team, err := gd.GetTeamById(2)
    if err != nil {
        fmt.Println(err)
    }
*/
func (gd *Godradis) GetTeamById(id int) (Team, error) {
	resp, err := gd.sendRequest("GET", fmt.Sprintf("teams/%v", id), nil)
	if err != nil {
		return Team{}, err
	}
	defer resp.Body.Close()
	var team Team
	if resp.StatusCode != http.StatusOK {
		return Team{}, errors.New("could not get team")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Team{}, err
	}

	err = json.Unmarshal(body, &team)
	if err != nil {
		return Team{}, err
	}
	return team, nil
}

/*
GetTeamByName searches for and returns a Team object based on the name. GetTeamByName works by calling GetAllTeams
first and then ranges over them comparing the name strings.

    gd := godradis.Godradis{}

    [...]

    team, err := gd.GetTeamByName("Test Client")
    if err != nil {
        fmt.Println(err)
    }
 */
func (gd *Godradis) GetTeamByName(name string) (Team, error) {
	teams, err := gd.GetAllTeams()
	if err != nil {
		return Team{}, err
	}
	for _, team := range teams {
		if strings.ToLower(team.Name) == strings.ToLower(name) {
			return team, nil
		}
	}
	return Team{}, errors.New(fmt.Sprintf("could not find team with name %s", name))
}

type teamDetails struct {
	Name string `json:"name,omitempty"`
	TeamSince string `json:"team_since,omitempty"`
}

func (td *teamDetails) parseArguments(name, teamSince interface{}) {
	if name == nil {
		td.Name = ""
	} else {
		td.Name = name.(string)
	}
	if teamSince == nil {
		td.TeamSince = ""
	} else {
		td.TeamSince = teamSince.(string)
	}
}

/*
CreateTeam takes a name and optional teamSince (in the form "YYYY-MM-DD") string and creates a new Team on the server,
returning a new Team object. teamSince defaults to the current date if it's not passed as an argument.

    gd := godradis.Godradis{}

    [...]

    team, _ := gd.CreateTeam("New Team", "2019-01-01")
    otherTeam, _ := gd.CreateTeam("Other Team")
 */
func (gd *Godradis) CreateTeam(name string, teamSince ...string) (Team, error) {
	// Required so that json.Marshal() sends the fields wrapped in a team{} json object
	type reqModel struct {
		TeamDetails teamDetails `json:"team"`
	}

	td := teamDetails{}
	td.Name = name
	if len(teamSince) > 0 {
		td.TeamSince = teamSince[0]
	}
	jsonBody, err := json.Marshal(&reqModel{td})
	if err != nil {
		return Team{}, err
	}
	resp, err := gd.sendRequest("POST", "teams", jsonBody)
	if err != nil {
		return Team{}, err
	}
	defer resp.Body.Close()
	var newTeam Team
	if resp.StatusCode != http.StatusCreated {
		return Team{}, errors.New("could not create team")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Team{}, err
	}

	err = json.Unmarshal(body, &newTeam)
	if err != nil {
		return Team{}, err
	}
	return newTeam, nil
}

/*
UpdateTeam takes a reference to an existing Team object and a name and teamSince (in the form "YYYY-MM-DD") string as
optional arguments. The Team argument is updated in-place.

    gd := godradis.Godradis{}

    [...]

    team, _ := gd.GetTeamByName("Test Client")
    err := gd.UpdateTeam(&team, nil, "2019-02-01")
 */
func (gd *Godradis) UpdateTeam(t *Team, name, teamSince interface{}) error {
	// Required so that json.Marshal() sends the fields wrapped in a team{} json object
	type reqModel struct {
		TeamDetails teamDetails `json:"team"`
	}
	td := teamDetails{}
	td.parseArguments(name, teamSince)
	jsonBody, err := json.Marshal(&reqModel{td})
	if err != nil {
		return err
	}
	resp, err := gd.sendRequest("PUT", fmt.Sprintf("teams/%v", t.Id), jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("could not update team")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &t)
	if err != nil {
		return err
	}
	return nil
}

/*
DeleteTeam takes a reference to a Team object and deletes the team on the server.

    gd := godradis.Godradis{}

    [...]

    team, _ := gd.CreateTeam("New Team", "2019-01-01")
    err := gd.DeleteTeam(&team)
    if err != nil {
        fmt.Println(err)
    }
 */
func (gd *Godradis) DeleteTeam(t *Team) error {
	resp, err := gd.sendRequest("DELETE", fmt.Sprintf("teams/%v", t.Id), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return errors.New("could not delete team")
	}
}

// Nodes endpoint

/*
GetAllNodes takes a reference to a Project object and returns a list of all Nodes that exist on the server for that project.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.GetProjectByName("Foobar External Network Penetration Test")
    nodes, _ := gd.GetAllNodes(&project)
 */
func (gd *Godradis) GetAllNodes(project *Project) ([]Node, error) {
	resp, err := gd.sendRequestWithProjectId("GET", "nodes", project.Id, nil)
	if err != nil {
		return []Node{}, err
	}
	defer resp.Body.Close()
	var nodes []Node
	if resp.StatusCode != http.StatusOK {
		return []Node{}, errors.New("could not get nodes list")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []Node{}, err
	}

	err = json.Unmarshal(body, &nodes)
	if err != nil {
		return []Node{}, err
	}
	for i := 0; i < len(nodes); i++ {
		nodes[i].Project = project
		nodes[i].setEvidenceNodeReferences()
		nodes[i].setNoteNodeReferences()
	}
	return nodes, nil
}

/*
GetNodeById takes a reference to a Project object and int id and returns the node associated with that id.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.GetProjectByName("Foobar External Network Penetration Test")
    node, _ := gd.GetNodeById(&project, 7)
 */
func (gd *Godradis) GetNodeById(project *Project, id int) (Node, error) {
	resp, err := gd.sendRequestWithProjectId("GET", fmt.Sprintf("nodes/%v", id), project.Id, nil)
	if err != nil {
		return Node{}, err
	}
	defer resp.Body.Close()
	var node Node
	if resp.StatusCode != http.StatusOK {
		return Node{}, errors.New("could not get node")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Node{}, err
	}

	err = json.Unmarshal(body, &node)
	if err != nil {
		return Node{}, err
	}
	node.Project = project
	node.setEvidenceNodeReferences()
	node.setNoteNodeReferences()
	return node, nil
}

/*
GetNodeByLabel searches for and returns a Node object based on the label. GetNodeByLabel works by calling GetAllNodes
first and then ranges over them comparing the label strings.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.GetProjectByName("Foobar External Network Penetration Test")
    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
 */
func (gd *Godradis) GetNodeByLabel(project *Project, label string) (Node, error) {
	nodes, err := gd.GetAllNodes(project)
	if err != nil {
		return Node{}, err
	}
	for _, node := range nodes {
		if strings.ToLower(node.Label) == strings.ToLower(label) {
			return node, nil
		}
	}
	return Node{}, errors.New(fmt.Sprintf("could not find node with label %s", label))
}

type nodeDetails struct {
	Label string `json:"label,omitempty"`
	TypeId int `json:"type_id,omitempty"`
	ParentId int `json:"parent_id,omitempty"`
	Position int `json:"position,omitempty"`
}

func (nd *nodeDetails) parseArguments(label, typeId, parentId, position interface{}) {
	if label == nil {
		nd.Label = ""
	} else {
		nd.Label = label.(string)
	}
	if typeId == nil {
		nd.TypeId = 0
	} else {
		nd.TypeId = typeId.(int)
	}
	if parentId == nil {
		nd.ParentId = 0
	} else {
		nd.ParentId = parentId.(int)
	}
	if position == nil {
		nd.Position = 0
	} else {
		nd.Position = position.(int)
	}
}

/*
CreateNode takes a reference to a Project object and several mandatory properties and creates a new Node on the server
and returns it. label is a string representing the name of the node. typeId is an int and can be 0 (a "default" node) or
1 (a "host" node). parentId is an int indicating the ID of the parent node if there is one, otherwise it will be created
as a top-level node. position is an int that determines where to insert the node within the existing node structure.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.GetProjectByName("Foobar External Network Penetration Test")
    node, _ := gd.CreateNode(&project, "127.0.0.1", 1, 14, 3)
 */
func (gd *Godradis) CreateNode(project *Project, label string, typeId int, parentId int, position int) (Node, error) {
	// BUG(njfox): The parentId argument to CreateNode may not be correctly serialized in the API request

	// Required so that json.Marshal() sends the fields wrapped in a node{} json object
	type reqModel struct {
		Node nodeDetails `json:"node"`
	}

	nd := nodeDetails{label, typeId, parentId, position}
	jsonBody, err := json.Marshal(&reqModel{nd})
	if err != nil {
		return Node{}, err
	}
	resp, err := gd.sendRequestWithProjectId("POST", "nodes", project.Id, jsonBody)
	if err != nil {
		return Node{}, err
	}
	defer resp.Body.Close()
	var newNode Node
	if resp.StatusCode != http.StatusCreated {
		return Node{}, errors.New("could not create node")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Node{}, err
	}

	err = json.Unmarshal(body, &newNode)
	if err != nil {
		return Node{}, err
	}
	newNode.Project = project
	newNode.setEvidenceNodeReferences()
	return newNode, nil
}

/*
UpdateNode takes a reference to an existing Node object and updates any non-nil properties passed to it as arguments.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.GetProjectByName("Foobar External Network Penetration Test")
    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    _ := gd.UpdateNode(&node, "localhost", nil, nil, nil)
 */
func (gd *Godradis) UpdateNode(n *Node, label, typeId, parentId, position interface{}) error {
	// Required so that json.Marshal() sends the fields wrapped in a node{} json object
	type reqModel struct {
		NodeDetails nodeDetails `json:"node"`
	}
	nd := nodeDetails{}
	nd.parseArguments(label, typeId, parentId, position)
	jsonBody, err := json.Marshal(&reqModel{nd})
	if err != nil {
		return err
	}
	resp, err := gd.sendRequestWithProjectId("PUT", fmt.Sprintf("nodes/%v", n.Id), n.Project.Id, jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("could not update node")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &n)
	if err != nil {
		return err
	}
	return nil
}

/*
DeleteNode takes a reference to an existing Node object and deletes it on the server.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.GetProjectByName("Foobar External Network Penetration Test")
    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    _ := gd.DeleteNode(&node)
 */
func (gd *Godradis) DeleteNode(n *Node) error {
	resp, err := gd.sendRequestWithProjectId("DELETE", fmt.Sprintf("nodes/%v", n.Id), n.Project.Id, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return errors.New("could not delete node")
	}
}

// Issues endpoint

/*
GetAllIssues takes a reference to a Project object and returns a list of all Issues that exist on the server for that project.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.GetProjectByName("Foobar External Network Penetration Test")
    issues, _ := gd.GetAllIssues(&project)
*/
func (gd *Godradis) GetAllIssues(project *Project) ([]Issue, error) {
	resp, err := gd.sendRequestWithProjectId("GET", "issues", project.Id, nil)
	if err != nil {
		return []Issue{}, err
	}
	defer resp.Body.Close()
	var issues []Issue
	if resp.StatusCode != http.StatusOK {
		return []Issue{}, errors.New("could not get issue list")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []Issue{}, err
	}

	err = json.Unmarshal(body, &issues)
	if err != nil {
		return []Issue{}, err
	}
	for i := 0; i < len(issues); i++ {
		issues[i].Project = project
	}
	return issues, nil
}

/*
GetIssueById takes a reference to a Project object and int id and returns the Issue associated with that id.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.GetProjectByName("Foobar External Network Penetration Test")
    issue, _ := gd.GetIssueById(&project, 12)
 */
func (gd *Godradis) GetIssueById(project *Project, id int) (Issue, error) {
	resp, err := gd.sendRequestWithProjectId("GET", fmt.Sprintf("issues/%v", id), project.Id, nil)
	if err != nil {
		return Issue{}, err
	}
	defer resp.Body.Close()
	var issue Issue
	if resp.StatusCode != http.StatusOK {
		return Issue{}, errors.New("could not get issue")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Issue{}, err
	}

	err = json.Unmarshal(body, &issue)
	if err != nil {
		return Issue{}, err
	}
	issue.Project = project
	return issue, nil
}


/*
GetIssueByTitle searches for and returns an Issue object based on the title. GetIssueByTitle works by calling GetAllIssues
first and then ranges over them comparing the title strings.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.GetProjectByName("Foobar External Network Penetration Test")
    issue, _ := gd.GetIssueByTitle(&project, "Cross-Site Scripting")
 */
func (gd *Godradis) GetIssueByTitle(project *Project, title string) (Issue, error) {
	issues, err := gd.GetAllIssues(project)
	if err != nil {
		return Issue{}, err
	}
	for _, issue := range issues {
		if strings.ToLower(issue.Title) == strings.ToLower(title) {
			return issue, nil
		}
	}
	return Issue{}, errors.New(fmt.Sprintf("could not find issue with title %s", title))
}

/*
CreateIssue takes a reference to a Project object and an OrderedMap containing the fields in the Issue body, creates a
new Issue on the server, and returns it.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.GetProjectByName("Foobar External Network Penetration Test")
    fields := orderedmap.New()
    fields.Set("Title", "Insecure Password Storage")
    fields.Set("Severity", "High")
    fields.Set("Finding Information", "Lorem ipsum dolor sit amet")
    issue, _ := gd.CreateIssue(&project, fields)
 */
func (gd *Godradis) CreateIssue(project *Project, fields *orderedmap.OrderedMap) (Issue, error) {
	text := parseOrderedMapFields(fields)
	issue, err := gd.CreateIssueFromText(project, text)
	if err != nil {
		return Issue{}, err
	}
	issue.Project = project
	return issue, nil
}

/*
CreateIssueFromText provides an alternate method for creating issues directly from a text string as opposed to the
OrderedMap approach used by CreateIssue. CreateIssueFromText takes a reference to a Project object and a string containing
the raw body content and returns the newly created Issue.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.GetProjectByName("Foobar External Network Penetration Test")
    issue, _ = gd.CreateIssueFromText(&project, "#[Title]#\r\nInsecure Password Storage\r\n\r\n#[Severity]#\r\nHigh")
 */
func (gd *Godradis) CreateIssueFromText(project *Project, text string) (Issue, error) {
	// Required so that json.Marshal() sends the fields wrapped in a issue{} json object
	type issueDetails struct {
		Text string `json:"text"`
	}
	type reqModel struct {
		IssueDetails issueDetails `json:"issue"`
	}

	jsonBody, err := json.Marshal(&reqModel{issueDetails{text}})
	if err != nil {
		return Issue{}, err
	}
	resp, err := gd.sendRequestWithProjectId("POST", "issues", project.Id, jsonBody)
	if err != nil {
		return Issue{}, err
	}
	defer resp.Body.Close()
	var newIssue Issue
	if resp.StatusCode != http.StatusCreated {
		return Issue{}, errors.New("could not create issue")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Issue{}, err
	}

	err = json.Unmarshal(body, &newIssue)
	if err != nil {
		return Issue{}, err
	}
	newIssue.Project = project
	return newIssue, nil
}

/*
UpdateIssue takes a reference to an existing Issue object and an OrderedMap containing the fields making up the content
of the Issue body, updates the Issue on the server, and modifies the local Issue object in place with the updated information.
Note that due to the way the Dradis API works, all fields in the body must be passed in the OrderedMap, not just the fields
that are being modified.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.GetProjectByName("Foobar External Network Penetration Test")
    issue, _ := gd.GetIssueByTitle(&project, "Insecure Password Storage")
    fields := issue.Fields
    fields.Set("Severity", "Medium")
    _ := gd.UpdateIssue(&issue, fields)
 */
func (gd *Godradis) UpdateIssue(issue *Issue, fields *orderedmap.OrderedMap) error {
	text := parseOrderedMapFields(fields)
	err := gd.UpdateIssueFromText(issue, text)
	if err != nil {
		return err
	}
	return nil
}

/*
UpdateIssueFromText provides an alternate method for updating issues directly from a text string as opposed to the
OrderedMap approach used by UpdateIssue. UpdateIssueFromText takes a reference to an existing Issue object and a string
containing the raw body content and modifies the existing Issue object in place. Note that due to the way the Dradis API
works, all fields in the body must be passed in the string, not just the fields that are being modified.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.GetProjectByName("Foobar External Network Penetration Test")
    issue, _ := gd.GetIssueByTitle(&project, "Insecure Password Storage")
    _ := gd.UpdateIssueFromText(&issue, "#[Title]#\r\nInsecure Password Storage\r\n\r\n#[Severity]#\r\Medium")
 */
func (gd *Godradis) UpdateIssueFromText(issue *Issue, text string) error {
	// Required so that json.Marshal() sends the fields wrapped in a issue{} json object
	type issueDetails struct {
		Text string `json:"text"`
	}
	type reqModel struct {
		IssueDetails issueDetails `json:"issue"`
	}

	jsonBody, err := json.Marshal(&reqModel{issueDetails{text}})
	if err != nil {
		return err
	}
	resp, err := gd.sendRequestWithProjectId("PUT", fmt.Sprintf("issues/%v", issue.Id), issue.Project.Id, jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("could not update issue")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &issue)
	if err != nil {
		return err
	}
	return nil
}

/*
DeleteIssue takes a reference to an existing Issue object and deletes it on the server.

    gd := godradis.Godradis{}

    [...]

    project, _ := gd.GetProjectByName("Foobar External Network Penetration Test")
    issue, _ := gd.GetIssueByTitle(&project, "Cross-Site Scripting")
    _ := gd.DeleteIssue(&issue)
 */
func (gd *Godradis) DeleteIssue(i *Issue) error {
	resp, err := gd.sendRequestWithProjectId("DELETE", fmt.Sprintf("issues/%v", i.Id), i.Project.Id, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return errors.New("could not delete issue")
	}
}

// Evidence endpoint

/*
GetAllEvidence takes a reference to a Node object and returns a list of all Evidence instances exist on the server for
that node.

    gd := godradis.Godradis{}

    [...]

    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    evidences, _ := gd.GetAllEvidence(&node)
 */
func (gd *Godradis) GetAllEvidence(node *Node) ([]Evidence, error) {
	resp, err := gd.sendRequestWithProjectId("GET", fmt.Sprintf("nodes/%v/evidence", node.Id), node.Project.Id, nil)
	if err != nil {
		return []Evidence{}, err
	}
	defer resp.Body.Close()
	var evidences []Evidence
	if resp.StatusCode != http.StatusOK {
		return []Evidence{}, errors.New("could not get evidence list")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []Evidence{}, err
	}

	err = json.Unmarshal(body, &evidences)
	if err != nil {
		return []Evidence{}, err
	}
	for i := 0; i < len(evidences); i++ {
		evidences[i].Node = node
	}
	return evidences, nil
}

/*
GetEvidenceById takes a reference to a Node object and int id and returns the Evidence instance associated with that id.

    gd := godradis.Godradis{}

    [...]

    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    evidence, _ := gd.GetEvidenceById(&node, 7)
 */
func (gd *Godradis) GetEvidenceById(node *Node, id int) (Evidence, error) {
	resp, err := gd.sendRequestWithProjectId("GET", fmt.Sprintf("nodes/%v/evidence/%v", node.Id, id), node.Project.Id, nil)
	if err != nil {
		return Evidence{}, err
	}
	defer resp.Body.Close()
	var evidence Evidence
	if resp.StatusCode != http.StatusOK {
		return Evidence{}, errors.New("could not get evidence")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Evidence{}, err
	}

	err = json.Unmarshal(body, &evidence)
	if err != nil {
		return Evidence{}, err
	}
	evidence.Node = node
	return evidence, nil
}

/*
CreateEvidence takes references to existing Node and Issue objects, and an OrderedMap object containing the content of the
Evidence instance. The Evidence is attached to the node and issue on the Dradis server and a local Evidence object is
returned.

    gd := godradis.Godradis{}

    [...]

    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    issue, _ := gd.GetIssueByTitle(&project, "Cross-Site Scripting")
    content := orderedmap.New()
    content.Set("Port", "443/tcp")
    content.Set("Reportable", "True")
    content.Set("Details", "Lorem ipsum dolor sit amet")
    evidence, _ := gd.CreateEvidence(&node, &issue, content)
 */
func (gd *Godradis) CreateEvidence(node *Node, issue *Issue, content *orderedmap.OrderedMap) (Evidence, error) {
	text := parseOrderedMapFields(content)
	evidence, err := gd.CreateEvidenceFromText(node, issue, text)
	if err != nil {
		return Evidence{}, err
	}
	return evidence, nil
}

/*
CreateEvidenceFromText provides an alternate method for creating evidence directly from a text string as opposed to the
OrderedMap approach used by CreateEvidence. CreateEvidenceFromText takes references to Node and Issue objects and a
string containing the raw body content and returns the newly created Evidence instance.

    gd := godradis.Godradis{}

    [...]

    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    issue, _ := gd.GetIssueByTitle(&project, "Cross-Site Scripting")
    evidence, _ := gd.CreateEvidence(&node, &issue, "#[Port]#\r\n443/tcp\r\n\r\n#[Details]#\r\nLorem ipsum dolor\r\n\r\n")
 */
func (gd *Godradis) CreateEvidenceFromText(node *Node, issue *Issue, content string) (Evidence, error) {
	// Required so that json.Marshal() sends the fields wrapped in an evidence{} json object
	type evidenceDetails struct {
		Content string `json:"content"`
		IssueId string `json:"issue_id"`
	}
	type reqModel struct {
		EvidenceDetails evidenceDetails `json:"evidence"`
	}

	jsonBody, err := json.Marshal(&reqModel{evidenceDetails{content, strconv.Itoa(issue.Id)}})
	if err != nil {
		return Evidence{}, err
	}
	resp, err := gd.sendRequestWithProjectId("POST", fmt.Sprintf("nodes/%v/evidence", node.Id), node.Project.Id, jsonBody)
	if err != nil {
		return Evidence{}, err
	}
	defer resp.Body.Close()
	var newEvidence Evidence
	if resp.StatusCode != http.StatusCreated {
		return Evidence{}, errors.New("could not create evidence")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Evidence{}, err
	}

	err = json.Unmarshal(body, &newEvidence)
	if err != nil {
		return Evidence{}, err
	}
	newEvidence.Node = node
	node.addEvidence(newEvidence)
	return newEvidence, nil
}

/*
UpdateEvidence takes a reference to an existing Evidence object, an OrderedMap containing the fields making up the content
of the Evidence body, and optionally a reference to an Issue object if the evidence is going to be attached to a different
issue on the server. UpdateEvidence updates the evidence on the server and modifies the local Evidence object in place with
the updated information. Note that due to the way the Dradis API works, all fields in the body must be passed in the OrderedMap,
not just the fields that are being modified.

    gd := godradis.Godradis{}

    [...]

    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    evidence, _ := gd.GetEvidenceById(&node, 2)
    newFields := evidence.CopyFields()
    newFields.Set("Port", "995/tcp")
    _ := gd.UpdateEvidence(&evidence, newFields)
 */
func (gd *Godradis) UpdateEvidence(evidence *Evidence, fields *orderedmap.OrderedMap, issue ...*Issue) error {
	text := parseOrderedMapFields(fields)
	var err error
	if len(issue) > 0 {
		err = gd.UpdateEvidenceFromText(evidence, text, issue[0])
	} else {
		err = gd.UpdateEvidenceFromText(evidence, text)
	}
	if err != nil {
		return err
	}
	return nil
}

/*
UpdateEvidenceFromText provides an alternate method for updating evidence directly from a text string as opposed to the
OrderedMap approach used by UpdateEvidence. UpdateEvidenceFromText takes a reference to an Evidence object, a string containing
the content, and optionally a reference to an Issue object if the evidence is being attached to a different issue. The
evidence object is modified in place with the updated information.

    gd := godradis.Godradis{}

    [...]

    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    evidence, _ := gd.GetEvidenceById(&node, 4)
    _ := gd.UpdateEvidenceFromText(&evidence, "#[Port]#\r\n443/tcp\r\n\r\n#[Details]#\r\nLorem ipsum dolor\r\n\r\n")
 */
func (gd *Godradis) UpdateEvidenceFromText(evidence *Evidence, content string, issue ...*Issue) error {
	// Required so that json.Marshal() sends the fields wrapped in a evidence{} json object
	type evidenceDetails struct {
		Content string `json:"content"`
		IssueId string `json:"issue_id,omitempty"`
	}
	type reqModel struct {
		EvidenceDetails evidenceDetails `json:"evidence"`
	}

	ed := evidenceDetails{}
	ed.Content = content
	if len(issue) > 0 {
		ed.IssueId = strconv.Itoa(issue[0].Id)
	}
	jsonBody, err := json.Marshal(&reqModel{ed})
	if err != nil {
		return err
	}
	resp, err := gd.sendRequestWithProjectId("PUT", fmt.Sprintf("nodes/%v/evidence/%v", evidence.Node.Id, evidence.Id), evidence.Node.Project.Id, jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("could not update evidence")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &evidence)
	if err != nil {
		return err
	}
	return nil
}

/*
DeleteEvidence takes a reference to an existing Evidence object and deletes it on the server.

    gd := godradis.Godradis{}

    [...]

    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    evidence, _ := gd.GetEvidenceById(&node, 4)
    _ := gd.DeleteEvidence(&evidence)
 */
func (gd *Godradis) DeleteEvidence(evidence *Evidence) error {
	resp, err := gd.sendRequestWithProjectId("DELETE", fmt.Sprintf("nodes/%v/evidence/%v", evidence.Node.Id, evidence.Id), evidence.Node.Project.Id, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		if evidence.Node != nil {
			evidence.Node.deleteEvidence(*evidence)
		}
		return nil
	} else {
		return errors.New("could not delete evidence")
	}
}

// Notes endpoint

/*
GetAllNotes takes a reference to a Node object and returns a list of all Notes attached to that node on the server.

    gd := godradis.Godradis{}

    [...]

    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    notes, _ := gd.GetAllNotes(&node)
 */
func (gd *Godradis) GetAllNotes(node *Node) ([]Note, error) {
	resp, err := gd.sendRequestWithProjectId("GET", fmt.Sprintf("nodes/%v/notes", node.Id), node.Project.Id, nil)
	if err != nil {
		return []Note{}, err
	}
	defer resp.Body.Close()
	var notes []Note
	if resp.StatusCode != http.StatusOK {
		return []Note{}, errors.New("could not get note list")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []Note{}, err
	}

	err = json.Unmarshal(body, &notes)
	if err != nil {
		return []Note{}, err
	}
	for i := 0; i < len(notes); i++ {
		notes[i].Node = node
	}
	return notes, nil
}

/*
GetNoteById takes a reference to a Node object and int id and returns the Note instance associated with that id.

    gd := godradis.Godradis{}

    [...]

    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    note, _ := gd.GetNoteById(&node, 7)
 */
func (gd *Godradis) GetNoteById(node *Node, id int) (Note, error) {
	resp, err := gd.sendRequestWithProjectId("GET", fmt.Sprintf("nodes/%v/notes/%v", node.Id, id), node.Project.Id, nil)
	if err != nil {
		return Note{}, err
	}
	defer resp.Body.Close()
	var note Note
	if resp.StatusCode != http.StatusOK {
		return Note{}, errors.New("could not get note from server")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Note{}, err
	}

	err = json.Unmarshal(body, &note)
	if err != nil {
		return Note{}, err
	}
	note.Node = node
	return note, nil
}

/*
GetNoteByTitle takes a reference to a Node object and string title and returns the first Note instance associated with that
title.

    gd := godradis.Godradis{}

    [...]

    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    note, _ := gd.GetNoteByTitle(&node, "Nmap Host Info")
 */
func (gd *Godradis) GetNoteByTitle(node *Node, title string) (Note, error) {
	notes, err := gd.GetAllNotes(node)
	if err != nil {
		return Note{}, err
	}
	for _, note := range notes {
		if strings.ToLower(note.Title) == strings.ToLower(title) {
			return note, nil
		}
	}
	return Note{}, errors.New(fmt.Sprintf("could not find note with title %s", title))
}

/*
CreateNote takes a reference to an existing Node object, an OrderedMap object containing the content of the
Note, and an optional integer category ID that sets the note category (Defaults to "Default Category" in Dradis). The Note
is attached to the node on the Dradis server and a local Note object is returned.

    gd := godradis.Godradis{}

    [...]

    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    fields := orderedmap.New()
    fields.Set("Hostnames", "foo.com\r\nexample.foo.com")
    note, _ := gd.CreateNote(&node, fields)
 */
func (gd *Godradis) CreateNote(node *Node, fields *orderedmap.OrderedMap, categoryId ...int) (Note, error) {
	text := parseOrderedMapFields(fields)
	var cid int
	if len(categoryId) > 0 {
		cid = categoryId[0]
	} else {
		cid = 6
	}
	note, err := gd.CreateNoteFromText(node, text, cid)
	if err != nil {
		return Note{}, err
	}
	return note, nil
}

/*
CreateNoteFromText takes a reference to an existing Node object, a string containing the body of the Note, and an optional
integer category ID that sets the note category (Defaults to "Default Category" in Dradis). The Note is attached to the
node on the Dradis server and a local Note object is returned.

    gd := godradis.Godradis{}

    [...]

    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    text := "#[Hostnames]\r\nfoo.com\r\nexample.foo.com\r\n\r\n#
    note, _ := gd.CreateNote(&node, text)
 */
func (gd *Godradis) CreateNoteFromText(node *Node, text string, categoryId ...int) (Note, error) {
	// Required so that json.Marshal() sends the fields wrapped in an note{} json object
	type noteDetails struct {
		Text string `json:"text"`
		CategoryId string `json:"category_id"`
	}
	type reqModel struct {
		NoteDetails noteDetails `json:"note"`
	}

	nd := noteDetails{}
	nd.Text = text
	if len(categoryId) > 0 {
		nd.CategoryId = strconv.Itoa(categoryId[0])
	} else {
		nd.CategoryId = "6" // Set category to "Default category"
	}
	jsonBody, err := json.Marshal(&reqModel{nd})
	if err != nil {
		return Note{}, err
	}
	resp, err := gd.sendRequestWithProjectId("POST", fmt.Sprintf("nodes/%v/notes", node.Id), node.Project.Id, jsonBody)
	if err != nil {
		return Note{}, err
	}
	defer resp.Body.Close()
	var newNote Note
	if resp.StatusCode != http.StatusCreated {
		return Note{}, errors.New("could not create note")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Note{}, err
	}

	err = json.Unmarshal(body, &newNote)
	if err != nil {
		return Note{}, err
	}
	newNote.Node = node
	node.addNote(newNote)
	return newNote, nil
}

/*
UpdateNote takes a reference to an existing Note object, an OrderedMap containing the fields making up the content
of the Note body, and an optional integer category ID that sets the note category (Defaults to "Default Category" in Dradis).
UpdateNote updates the note on the server and modifies the local Note object in place with the updated information. Note
that due to the way the Dradis API works, all fields in the body must be passed in the OrderedMap, not just the fields that
are being modified.

    gd := godradis.Godradis{}

    [...]

    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    note, _ := gd.GetNoteByTitle(&node, "Nmap Host Info")
    newFields := note.CopyFields()
    newFields.Set("Hostnames", "sub.foo.com")
    _ := gd.UpdateNote(&note, newFields)
 */
func (gd *Godradis) UpdateNote(note *Note, fields *orderedmap.OrderedMap, categoryId ...int) error {
	text := parseOrderedMapFields(fields)
	var err error
	if len(categoryId) > 0 {
		err = gd.UpdateNoteFromText(note, text, categoryId[0])
	} else {
		err = gd.UpdateNoteFromText(note, text)
	}

	if err != nil {
		return err
	}
	return nil
}

/*
UpdateNoteFromText takes a reference to an existing Note object, a string containing the body of the Note, and an optional
integer category ID that sets the note category (Defaults to "Default Category" in Dradis). UpdateNoteFromText updates the
note on the server and modifies the local Note object in place with the updated information.

    gd := godradis.Godradis{}

    [...]

    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    note, _ := gd.GetNoteByTitle(&node, "Nmap Host Info")
    text := "#[Hostnames]#\r\nsub.foo.com\r\n\r\n
    note, _ := gd.UpdateNoteFromText(&node, text)
 */
func (gd *Godradis) UpdateNoteFromText(note *Note, text string, categoryId ...int) error {
	// Required so that json.Marshal() sends the fields wrapped in a note{} json object
	type noteDetails struct {
		Text string `json:"text,omitempty"`
		CategoryId string `json:"category_id,omitempty"`
	}
	type reqModel struct {
		NoteDetails noteDetails `json:"note"`
	}

	nd := noteDetails{}
	nd.Text = text
	if len(categoryId) > 0 {
		nd.CategoryId = strconv.Itoa(categoryId[0])
	}

	jsonBody, err := json.Marshal(&reqModel{nd})
	if err != nil {
		return err
	}
	resp, err := gd.sendRequestWithProjectId("PUT", fmt.Sprintf("nodes/%v/notes/%v", note.Node.Id, note.Id), note.Node.Project.Id, jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("could not update note")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &note)
	if err != nil {
		return err
	}
	return nil
}

/*
DeleteNote takes a reference to an existing Note object and deletes it on the server.

    gd := godradis.Godradis{}

    [...]

    node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
    note, _ := gd.GetNoteByTitle(&node, "Nmap Host Info")
    _ := gd.DeleteNote(&note)
 */
func (gd *Godradis) DeleteNote(note *Note) error {
	resp, err := gd.sendRequestWithProjectId("DELETE", fmt.Sprintf("nodes/%v/notes/%v", note.Node.Id, note.Id), note.Node.Project.Id, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		if note.Node != nil {
			note.Node.deleteNote(*note)
		}
		return nil
	} else {
		return errors.New("could not delete note")
	}
}

// Attachments endpoint

/*
GetAllAttachments takes a reference to an existing Node object and returns a slice of all attachments associated with that
node.
 */
func (gd *Godradis) GetAllAttachments(node *Node) ([]Attachment, error) {
	resp, err := gd.sendRequestWithProjectId("GET", fmt.Sprintf("nodes/%v/attachments", node.Id), node.Project.Id, nil)
	if err != nil {
		return []Attachment{}, err
	}
	defer resp.Body.Close()
	var attachments []Attachment
	if resp.StatusCode != http.StatusOK {
		return []Attachment{}, errors.New("could not get attachment list")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []Attachment{}, err
	}

	err = json.Unmarshal(body, &attachments)
	if err != nil {
		return []Attachment{}, err
	}
	for i := 0; i < len(attachments); i++ {
		attachments[i].Node = node
	}
	return attachments, nil
}

/*
GetAttachmentByName takes a reference to an existing Node object and a string filename and returns an Attachment object
if it is found on the server.
 */
func (gd *Godradis) GetAttachmentByName(node *Node, filename string) (Attachment, error) {
	escapedFilename := url.PathEscape(filename)
	resp, err := gd.sendRequestWithProjectId("GET", fmt.Sprintf("nodes/%v/attachments/%v", node.Id, escapedFilename), node.Project.Id, nil)
	if err != nil {
		return Attachment{}, err
	}
	defer resp.Body.Close()
	var attachment Attachment
	if resp.StatusCode != http.StatusOK {
		return Attachment{}, errors.New("could not get attachment")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Attachment{}, err
	}

	err = json.Unmarshal(body, &attachment)
	if err != nil {
		return Attachment{}, err
	}
	attachment.Node = node
	return attachment, nil
}

/*
UploadAttachments takes a reference to an existing Node object and a slice of strings containing filepaths and uploads
these attachments to the Dradis server. A slice of Attachment objects is returned.
 */
func (gd *Godradis) UploadAttachments(node *Node, filePath []string) ([]Attachment, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, path := range filePath {
		file, err := os.Open(path)
		if err != nil {
			return []Attachment{}, err
		}

		part, err := writer.CreateFormFile("files[]", filepath.Base(path))
		if err != nil {
			return []Attachment{}, err
		}
		_, err = io.Copy(part, file)
		file.Close()
	}
	err := writer.Close()
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/pro/api/nodes/%v/attachments", gd.Config.BaseUrl, node.Id), body)
	req.Header.Add("Authorization", fmt.Sprintf(`Token token="%s"`, gd.Config.ApiKey))
	req.Header.Set("Dradis-Project-Id", strconv.Itoa(node.Project.Id))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := gd.httpClient.Do(req)
	if err != nil {
		return []Attachment{}, err
	}
	defer resp.Body.Close()
	var attachments []Attachment
	if resp.StatusCode != http.StatusCreated {
		return []Attachment{}, errors.New("could not upload attachments")
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []Attachment{}, err
	}

	err = json.Unmarshal(respBody, &attachments)
	if err != nil {
		return []Attachment{}, err
	}
	for i := 0; i < len(attachments); i++ {
		attachments[i].Node = node
	}
	return attachments, nil
}

/*
DeleteAttachment takes a reference to an existing Attachment object and deletes it from the server. The local Attachment
object reference is set to nil.
 */
func (gd *Godradis) DeleteAttachment(attachment *Attachment) error {
	resp, err := gd.sendRequestWithProjectId("DELETE", fmt.Sprintf("nodes/%v/attachments/%v", attachment.Node.Id, attachment.Filename), attachment.Node.Project.Id, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return errors.New("could not delete attachment")
	}
}

// IssueLib endpoint

func (gd *Godradis) GetIssueLibrary() ([]IssueLib, error) {
	resp, err := gd.sendRequest("GET", "addons/issuelib/entries", nil)
	if err != nil {
		return []IssueLib{}, err
	}
	defer resp.Body.Close()
	var issueLibs []IssueLib
	if resp.StatusCode != http.StatusOK {
		return []IssueLib{}, errors.New("could not get issue library entries")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []IssueLib{}, err
	}

	err = json.Unmarshal(body, &issueLibs)
	if err != nil {
		return []IssueLib{}, err
	}
	return issueLibs, nil
}

func (gd *Godradis) GetIssueLibraryById(id int) (IssueLib, error) {
	resp, err := gd.sendRequest("GET", fmt.Sprintf("addons/issuelib/entries/%v", id), nil)
	if err != nil {
		return IssueLib{}, err
	}
	defer resp.Body.Close()
	var issueLib IssueLib
	if resp.StatusCode != http.StatusOK {
		return IssueLib{}, errors.New("could not get issue library entry")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return IssueLib{}, err
	}

	err = json.Unmarshal(body, &issueLib)
	if err != nil {
		return IssueLib{}, err
	}
	return issueLib, nil
}

func (gd *Godradis) DeleteIssueLibraryById(entry IssueLib) error {
	resp, err := gd.sendRequest("DELETE", fmt.Sprintf("addons/issuelib/entries/%v", entry.Id), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return errors.New("could not delete issue library entry")
	}
}