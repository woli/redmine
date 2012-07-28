package redmine

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type RelationType string

const (
	RELATION_TYPE_RELATES    RelationType = "relates"
	RELATION_TYPE_DUPLICATES RelationType = "duplicates"
	RELATION_TYPE_DUPLICATED RelationType = "duplicated"
	RELATION_TYPE_BLOCKS     RelationType = "blocks"
	RELATION_TYPE_BLOCKED    RelationType = "blocked"
	RELATION_TYPE_PRECEDES   RelationType = "precedes"
	RELATION_TYPE_FOLLOWS    RelationType = "follows"
)

type IssueRelation struct {
	Id           int
	IssueId      int
	IssueToId    int
	RelationType RelationType
	Delay        int
}

type issueRelationDec struct {
	Id           *int    `json:"id"`
	IssueId      *int    `json:"issue_id"`
	IssueToId    *int    `json:"issue_to_id"`
	RelationType *string `json:"relation_type"`
	Delay        *int    `json:"delay"`
}

type issueRelationEnc struct {
	IssueToId    int    `json:"issue_to_id,omitempty"`
	RelationType string `json:"relation_type,omitempty"`
	Delay        int    `json:"delay,omitempty"`
}

func (i *issueRelationDec) decode() *IssueRelation {
	return &IssueRelation{
		Id:           ptrToInt(i.Id),
		IssueId:      ptrToInt(i.IssueId),
		IssueToId:    ptrToInt(i.IssueToId),
		RelationType: RelationType(ptrToString(i.RelationType)),
		Delay:        ptrToInt(i.Delay),
	}
}

func (i *IssueRelation) encode() *issueRelationEnc {
	return &issueRelationEnc{
		IssueToId:    i.IssueToId,
		RelationType: string(i.RelationType),
		Delay:        i.Delay,
	}
}

func unmarshalIssueRelation(data []byte) (*IssueRelation, error) {
	type response struct {
		Relation *issueRelationDec `json:"relation"`
	}

	res := &response{}
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res.Relation.decode(), nil
}

func marshalIssueRelation(issueRelation *IssueRelation) ([]byte, error) {
	type request struct {
		Relation *issueRelationEnc `json:"relation"`
	}

	req := &request{issueRelation.encode()}
	return json.Marshal(req)
}

func (r *Redmine) GetIssueRelations(issueId int, v *url.Values) ([]*IssueRelation, *Pagination, error) {
	path := fmt.Sprintf("issues/%v/relations.json", issueId)
	data, err := r.get(path, v)
	if err != nil {
		return nil, nil, err
	}

	type response struct {
		Relations  []*issueRelationDec `json:"relations"`
		TotalCount int                 `json:"total_count"`
		Limit      int                 `json:"limit"`
		Offset     int                 `json:"offset"`
	}

	res := &response{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, err
	}

	issueRelations := make([]*IssueRelation, len(res.Relations))
	for i := range res.Relations {
		issueRelations[i] = res.Relations[i].decode()
	}

	pagination := &Pagination{res.TotalCount, res.Limit, res.Offset}
	return issueRelations, pagination, nil
}

func (r *Redmine) GetIssueRelation(id int) (*IssueRelation, error) {
	path := fmt.Sprintf("relations/%v.json", id)
	data, err := r.get(path, nil)
	if err != nil {
		return nil, err
	}

	return unmarshalIssueRelation(data)
}

func (r *Redmine) CreateIssueRelation(issueRelation *IssueRelation) (*IssueRelation, error) {
	body, err := marshalIssueRelation(issueRelation)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("issues/%v/relations.json", issueRelation.IssueId)
	data, err := r.post(path, body)
	if err != nil {
		return nil, err
	}

	return unmarshalIssueRelation(data)
}

func (r *Redmine) DeleteIssueRelation(id int) error {
	path := fmt.Sprintf("relations/%v.json", id)
	return r.delete(path)
}
