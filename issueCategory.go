package redmine

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

type IssueCategory struct {
	Id         int
	Name       string
	Project    *Project
	AssignedTo *User
}

type issueCategoryDec struct {
	Id         *int        `json:"id"`
	Name       *string     `json:"name"`
	Project    *projectDec `json:"project"`
	AssignedTo *userDec    `json:"assigned_to"`
}

type issueCategoryEnc struct {
	Name         string `json:"name,omitempty"`
	AssignedToId int    `json:"assigned_to_id,omitempty"`
}

func (i *issueCategoryDec) decode() *IssueCategory {
	dec := &IssueCategory{
		Id:   ptrToInt(i.Id),
		Name: ptrToString(i.Name),
	}
	if i.Project != nil {
		dec.Project = i.Project.decode()
	}
	if i.AssignedTo != nil {
		dec.AssignedTo = i.AssignedTo.decode()
	}

	return dec
}

func (i *IssueCategory) encode() *issueCategoryEnc {
	enc := &issueCategoryEnc{
		Name: i.Name,
	}

	if i.AssignedTo != nil {
		enc.AssignedToId = i.AssignedTo.Id
	}

	return enc
}

func unmarshalIssueCategory(data []byte) (*IssueCategory, error) {
	type response struct {
		IssueCategory *issueCategoryDec `json:"issue_category"`
	}

	res := &response{}
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res.IssueCategory.decode(), nil
}

func marshalIssueCategory(issueCategory *IssueCategory) ([]byte, error) {
	type request struct {
		IssueCategory *issueCategoryEnc `json:"issue_category"`
	}

	req := &request{issueCategory.encode()}
	return json.Marshal(req)
}

func (r *Redmine) GetIssueCategories(projectId int, v *url.Values) ([]*IssueCategory, *Pagination, error) {
	path := fmt.Sprintf("projects/%v/issue_categories.json", projectId)
	data, err := r.get(path, v)
	if err != nil {
		return nil, nil, err
	}

	type response struct {
		IssueCategories []*issueCategoryDec `json:"issue_categories"`
		TotalCount      int                 `json:"total_count"`
		Limit           int                 `json:"limit"`
		Offset          int                 `json:"offset"`
	}

	res := &response{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, err
	}

	issueCategories := make([]*IssueCategory, len(res.IssueCategories))
	for i := range res.IssueCategories {
		issueCategories[i] = res.IssueCategories[i].decode()
	}

	pagination := &Pagination{res.TotalCount, res.Limit, res.Offset}
	return issueCategories, pagination, nil
}

func (r *Redmine) GetIssueCategory(id int) (*IssueCategory, error) {
	path := fmt.Sprintf("issue_categories/%v.json", id)
	data, err := r.get(path, nil)
	if err != nil {
		return nil, err
	}

	return unmarshalIssueCategory(data)
}

func (r *Redmine) CreateIssueCategory(issueCategory *IssueCategory) (*IssueCategory, error) {
	if issueCategory.Project == nil {
		return nil, errors.New("Missing required field: Project")
	}

	body, err := marshalIssueCategory(issueCategory)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("projects/%v/issue_categories.json", issueCategory.Project.Id)
	data, err := r.post(path, body)
	if err != nil {
		return nil, err
	}

	return unmarshalIssueCategory(data)
}

func (r *Redmine) UpdateIssueCategory(issueCategory *IssueCategory) error {
	body, err := marshalIssueCategory(issueCategory)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("issue_categories/%v.json", issueCategory.Id)
	return r.put(path, body)
}

func (r *Redmine) DeleteIssueCategory(id int) error {
	path := fmt.Sprintf("issue_categories/%v.json", id)
	return r.delete(path)
}

func (r *Redmine) DeleteIssueCategoryAndReassign(id, reassignToId int) error {
	path := fmt.Sprintf("issue_categories/%v.json?reassign_to_id=%v", id, reassignToId)
	return r.delete(path)
}
