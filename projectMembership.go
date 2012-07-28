package redmine

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

type ProjectMembership struct {
	Id      int
	Project *Project
	Roles   []*Role
	Group   *Group
	User    *User
}

type projectMembershipDec struct {
	Id      *int        `json:"id"`
	Project *projectDec `json:"project"`
	Roles   []*roleDec  `json:"roles"`
	Group   *groupDec   `json:"group"`
	User    *userDec    `json:"user"`
}

type projectMembershipEnc struct {
	UserId  int   `json:"user_id,omitempty"`
	RoleIds []int `json:"role_ids,omitempty"`
}

func (p *projectMembershipDec) decode() *ProjectMembership {
	dec := &ProjectMembership{
		Id: ptrToInt(p.Id),
	}
	if p.Project != nil {
		dec.Project = p.Project.decode()
	}
	if p.Roles != nil {
		dec.Roles = make([]*Role, len(p.Roles))
		for i := range p.Roles {
			dec.Roles[i] = p.Roles[i].decode()
		}
	}
	if p.Group != nil {
		dec.Group = p.Group.decode()
	}
	if p.User != nil {
		dec.User = p.User.decode()
	}

	return dec
}

func (p *ProjectMembership) encode() *projectMembershipEnc {
	enc := &projectMembershipEnc{}
	if p.User != nil {
		enc.UserId = p.User.Id
	}
	if p.Roles != nil {
		roles := make([]int, len(p.Roles))
		for i := range p.Roles {
			roles[i] = p.Roles[i].Id
		}
		enc.RoleIds = roles
	}

	return enc
}

func unmarshalProjectMembership(data []byte) (*ProjectMembership, error) {
	type response struct {
		Membership *projectMembershipDec `json:"membership"`
	}

	res := &response{}
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res.Membership.decode(), nil
}

func marshalProjectMembership(projectMembership *ProjectMembership) ([]byte, error) {
	type request struct {
		Membership *projectMembershipEnc `json:"membership"`
	}

	req := &request{projectMembership.encode()}
	return json.Marshal(req)
}

func (r *Redmine) GetProjectMemberships(projectId int, v *url.Values) ([]*ProjectMembership, *Pagination, error) {
	path := fmt.Sprintf("projects/%v/memberships.json", projectId)
	data, err := r.get(path, v)
	if err != nil {
		return nil, nil, err
	}

	type response struct {
		Memberships []*projectMembershipDec `json:"memberships"`
		TotalCount  int                     `json:"total_count"`
		Limit       int                     `json:"limit"`
		Offset      int                     `json:"offset"`
	}

	res := &response{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, err
	}

	projectMemberships := make([]*ProjectMembership, len(res.Memberships))
	for i := range res.Memberships {
		projectMemberships[i] = res.Memberships[i].decode()
	}

	pagination := &Pagination{res.TotalCount, res.Limit, res.Offset}
	return projectMemberships, pagination, nil
}

func (r *Redmine) GetProjectMembership(id int) (*ProjectMembership, error) {
	path := fmt.Sprintf("memberships/%v.json", id)
	data, err := r.get(path, nil)
	if err != nil {
		return nil, err
	}

	return unmarshalProjectMembership(data)
}

func (r *Redmine) CreateProjectMembership(projectMembership *ProjectMembership) (*ProjectMembership, error) {
	if projectMembership.Project == nil {
		return nil, errors.New("Missing required field: Project")
	}

	body, err := marshalProjectMembership(projectMembership)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("projects/%v/memberships.json", projectMembership.Project.Id)
	data, err := r.post(path, body)
	if err != nil {
		return nil, err
	}

	return unmarshalProjectMembership(data)
}

func (r *Redmine) UpdateProjectMembership(projectMembership *ProjectMembership) error {
	body, err := marshalProjectMembership(projectMembership)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("memberships/%v.json", projectMembership.Id)
	return r.put(path, body)
}

func (r *Redmine) DeleteProjectMembership(id int) error {
	path := fmt.Sprintf("memberships/%v.json", id)
	return r.delete(path)
}
