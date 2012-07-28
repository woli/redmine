package redmine

import (
	"encoding/json"
	"net/url"
)

type Query struct {
	Id        int
	Name      string
	IsPublic  bool
	ProjectId int
}

type queryDec struct {
	Id        *int    `json:"id"`
	Name      *string `json:"name"`
	IsPublic  *bool   `json:"is_public"`
	ProjectId *int    `json:"project_id"`
}

func (q *queryDec) decode() *Query {
	return &Query{
		Id:        ptrToInt(q.Id),
		Name:      ptrToString(q.Name),
		IsPublic:  ptrToBool(q.IsPublic),
		ProjectId: ptrToInt(q.ProjectId),
	}
}

func (r *Redmine) GetQueries(v *url.Values) ([]*Query, *Pagination, error) {
	data, err := r.get("queries.json", v)
	if err != nil {
		return nil, nil, err
	}

	type response struct {
		Queries    []*queryDec `json:"queries"`
		TotalCount int         `json:"total_count"`
		Limit      int         `json:"limit"`
		Offset     int         `json:"offset"`
	}

	res := &response{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, err
	}

	queries := make([]*Query, len(res.Queries))
	for i := range res.Queries {
		queries[i] = res.Queries[i].decode()
	}

	pagination := &Pagination{res.TotalCount, res.Limit, res.Offset}
	return queries, pagination, nil
}
