package redmine

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type Project struct {
	Id              int
	Identifier      string
	Name            string
	Description     string
	Homepage        string
	Parent          *Project
	Trackers        []*Tracker
	IssueCategories []*IssueCategory
	CustomFields    []*CustomField
	CreatedOn       time.Time
	UpdatedOn       time.Time
}

type projectDec struct {
	Id              *int                `json:"id"`
	Identifier      *string             `json:"identifier"`
	Name            *string             `json:"name"`
	Description     *string             `json:"description"`
	Homepage        *string             `json:"homepage"`
	Parent          *projectDec         `json:"parent"`
	Trackers        []*trackerDec       `json:"trackers"`
	IssueCategories []*issueCategoryDec `json:"issue_categories"`
	CustomFields    []*customFieldDec   `json:"custom_fields"`
	CreatedOn       *string             `json:"created_on"`
	UpdatedOn       *string             `json:"updated_on"`
}

type projectEnc struct {
	Identifier        string                 `json:"identifier,omitempty"`
	Name              string                 `json:"name,omitempty"`
	Description       string                 `json:"description,omitempty"`
	Homepage          string                 `json:"homepage,omitempty"`
	ParentId          int                    `json:"parent_id,omitempty"`
	Trackers          []*Tracker             `json:"trackers,omitempty"`
	CustomFieldValues map[string]interface{} `json:"custom_field_values,omitempty"`
}

func (p *projectDec) decode() *Project {
	dec := &Project{
		Id:          ptrToInt(p.Id),
		Identifier:  ptrToString(p.Identifier),
		Name:        ptrToString(p.Name),
		Description: ptrToString(p.Description),
		Homepage:    ptrToString(p.Homepage),
		CreatedOn:   strToTime(TimeLayout, ptrToString(p.CreatedOn)),
		UpdatedOn:   strToTime(TimeLayout, ptrToString(p.UpdatedOn)),
	}
	if p.Trackers != nil {
		dec.Trackers = make([]*Tracker, len(p.Trackers))
		for i := range p.Trackers {
			dec.Trackers[i] = p.Trackers[i].decode()
		}
	}
	if p.Parent != nil {
		dec.Parent = p.Parent.decode()
	}
	if p.IssueCategories != nil {
		dec.IssueCategories = make([]*IssueCategory, len(p.IssueCategories))
		for i := range p.IssueCategories {
			dec.IssueCategories[i] = p.IssueCategories[i].decode()
		}
	}
	if p.CustomFields != nil {
		dec.CustomFields = make([]*CustomField, len(p.CustomFields))
		for i := range p.CustomFields {
			dec.CustomFields[i] = p.CustomFields[i].decode()
		}
	}

	return dec
}

func (p *Project) encode() *projectEnc {
	enc := &projectEnc{
		Identifier:  p.Identifier,
		Name:        p.Name,
		Description: p.Description,
		Homepage:    p.Homepage,
		Trackers:    p.Trackers,
	}
	if p.Parent != nil {
		enc.ParentId = p.Parent.Id
	}
	if p.CustomFields != nil {
		enc.CustomFieldValues = customFieldsToMap(p.CustomFields)
	}

	return enc
}

func unmarshalProject(data []byte) (*Project, error) {
	type response struct {
		Project *projectDec `json:"project"`
	}

	res := &response{}
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res.Project.decode(), nil
}

func marshalProject(project *Project) ([]byte, error) {
	type request struct {
		Project *projectEnc `json:"project"`
	}

	req := &request{project.encode()}
	return json.Marshal(req)
}

func (r *Redmine) GetProjects(v *url.Values) ([]*Project, *Pagination, error) {
	data, err := r.get("projects.json", v)
	if err != nil {
		return nil, nil, err
	}

	type response struct {
		Projects   []*projectDec `json:"projects"`
		TotalCount int           `json:"total_count"`
		Limit      int           `json:"limit"`
		Offset     int           `json:"offset"`
	}

	res := &response{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, err
	}

	projects := make([]*Project, len(res.Projects))
	for i := range res.Projects {
		projects[i] = res.Projects[i].decode()
	}

	pagination := &Pagination{res.TotalCount, res.Limit, res.Offset}
	return projects, pagination, nil
}

func (r *Redmine) GetProject(id int, inc *ProjectInclude) (*Project, error) {
	path := fmt.Sprintf("projects/%v.json", id)
	data, err := r.get(path, inc.values())
	if err != nil {
		return nil, err
	}

	return unmarshalProject(data)
}

func (r *Redmine) CreateProject(project *Project) (*Project, error) {
	body, err := marshalProject(project)
	if err != nil {
		return nil, err
	}

	data, err := r.post("projects.json", body)
	if err != nil {
		return nil, err
	}

	return unmarshalProject(data)
}

func (r *Redmine) UpdateProject(project *Project) error {
	body, err := marshalProject(project)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("projects/%v.json", project.Id)
	return r.put(path, body)
}

func (r *Redmine) DeleteProject(id int) error {
	path := fmt.Sprintf("projects/%v.json", id)
	return r.delete(path)
}

type ProjectInclude struct {
	Trackers        bool
	IssueCategories bool
}

func (p *ProjectInclude) values() *url.Values {
	m := map[string]bool{
		"trackers":         p.Trackers,
		"issue_categories": p.IssueCategories,
	}

	includes := make([]string, 0)
	for k, v := range m {
		if v {
			includes = append(includes, k)
		}
	}

	v := &url.Values{}
	if len(includes) > 0 {
		v.Add("include", strings.Join(includes, ","))
	}

	return v
}
