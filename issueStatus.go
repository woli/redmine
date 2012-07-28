package redmine

import (
	"encoding/json"
	"net/url"
)

type IssueStatus struct {
	Id        int
	Name      string
	IsDefault bool
	IsClosed  bool
}

type issueStatusDec struct {
	Id        *int    `json:"id"`
	Name      *string `json:"name"`
	IsDefault *bool   `json:"is_default"`
	IsClosed  *bool   `json:"is_closed"`
}

func (i *issueStatusDec) decode() *IssueStatus {
	return &IssueStatus{
		Id:        ptrToInt(i.Id),
		Name:      ptrToString(i.Name),
		IsDefault: ptrToBool(i.IsDefault),
		IsClosed:  ptrToBool(i.IsClosed),
	}
}

func (r *Redmine) GetIssueStatuses(v *url.Values) ([]*IssueStatus, *Pagination, error) {
	data, err := r.get("issue_statuses.json", v)
	if err != nil {
		return nil, nil, err
	}

	type response struct {
		IssueStatuses []*issueStatusDec `json:"issue_statuses"`
		TotalCount    int               `json:"total_count"`
		Limit         int               `json:"limit"`
		Offset        int               `json:"offset"`
	}

	res := &response{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, err
	}

	issueStatuses := make([]*IssueStatus, len(res.IssueStatuses))
	for i := range res.IssueStatuses {
		issueStatuses[i] = res.IssueStatuses[i].decode()
	}

	pagination := &Pagination{res.TotalCount, res.Limit, res.Offset}
	return issueStatuses, pagination, nil
}
