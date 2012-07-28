package redmine

import (
	"encoding/json"
	"net/url"
)

type Role struct {
	Id        int
	Name      string
	Inherited bool
}

type roleDec struct {
	Id        *int    `json:"id"`
	Name      *string `json:"name"`
	Inherited *bool   `json:"inherited"`
}

func (r *roleDec) decode() *Role {
	return &Role{
		Id:        ptrToInt(r.Id),
		Name:      ptrToString(r.Name),
		Inherited: ptrToBool(r.Inherited),
	}
}

func (r *Redmine) GetRoles(v *url.Values) ([]*Role, *Pagination, error) {
	data, err := r.get("roles.json", v)
	if err != nil {
		return nil, nil, err
	}

	type response struct {
		Roles      []*roleDec `json:"roles"`
		TotalCount int        `json:"total_count"`
		Limit      int        `json:"limit"`
		Offset     int        `json:"offset"`
	}

	res := &response{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, err
	}

	roles := make([]*Role, len(res.Roles))
	for i := range res.Roles {
		roles[i] = res.Roles[i].decode()
	}

	pagination := &Pagination{res.TotalCount, res.Limit, res.Offset}
	return roles, pagination, nil
}
