package redmine

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type Group struct {
	Id           int
	Name         string
	Users        []*User
	CustomFields []*CustomField
}

type groupDec struct {
	Id           *int              `json:"id"`
	Name         *string           `json:"name"`
	Users        []*userDec        `json:"users"`
	CustomFields []*customFieldDec `json:"custom_fields"`
}

type groupEnc struct {
	Name              string                 `json:"name,omitempty"`
	UserIds           []int                  `json:"user_ids,omitempty"`
	CustomFieldValues map[string]interface{} `json:"custom_field_values,omitempty"`
}

func (g *groupDec) decode() *Group {
	dec := &Group{
		Id:   ptrToInt(g.Id),
		Name: ptrToString(g.Name),
	}
	if g.Users != nil {
		dec.Users = make([]*User, len(g.Users))
		for i := range g.Users {
			dec.Users[i] = g.Users[i].decode()
		}
	}
	if g.CustomFields != nil {
		dec.CustomFields = make([]*CustomField, len(g.CustomFields))
		for i := range g.CustomFields {
			dec.CustomFields[i] = g.CustomFields[i].decode()
		}
	}

	return dec
}

func (g *Group) encode() *groupEnc {
	enc := &groupEnc{
		Name: g.Name,
	}
	if g.Users != nil {
		userIds := make([]int, len(g.Users))
		for i := range g.Users {
			userIds[i] = g.Users[i].Id
		}
		enc.UserIds = userIds
	}
	if g.CustomFields != nil {
		enc.CustomFieldValues = customFieldsToMap(g.CustomFields)
	}

	return enc
}

func unmarshalGroup(data []byte) (*Group, error) {
	type response struct {
		Group *groupDec `json:"group"`
	}

	res := &response{}
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res.Group.decode(), nil
}

func marshalGroup(group *Group) ([]byte, error) {
	type request struct {
		Group *groupEnc `json:"group"`
	}

	req := &request{group.encode()}
	return json.Marshal(req)
}

func (r *Redmine) GetGroups(v *url.Values) ([]*Group, *Pagination, error) {
	data, err := r.get("groups.json", v)
	if err != nil {
		return nil, nil, err
	}

	type response struct {
		Groups     []*groupDec `json:"groups"`
		TotalCount int         `json:"total_count"`
		Limit      int         `json:"limit"`
		Offset     int         `json:"offset"`
	}

	res := &response{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, err
	}

	groups := make([]*Group, len(res.Groups))
	for i := range res.Groups {
		groups[i] = res.Groups[i].decode()
	}

	pagination := &Pagination{res.TotalCount, res.Limit, res.Offset}
	return groups, pagination, nil
}

func (r *Redmine) GetGroup(id int, inc *GroupInclude) (*Group, error) {
	path := fmt.Sprintf("groups/%v.json", id)
	data, err := r.get(path, inc.values())
	if err != nil {
		return nil, err
	}

	return unmarshalGroup(data)
}

func (r *Redmine) CreateGroup(group *Group) (*Group, error) {
	body, err := marshalGroup(group)
	if err != nil {
		return nil, err
	}

	data, err := r.post("groups.json", body)
	if err != nil {
		return nil, err
	}

	return unmarshalGroup(data)
}

func (r *Redmine) UpdateGroup(group *Group) error {
	body, err := marshalGroup(group)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("groups/%v.json", group.Id)
	return r.put(path, body)
}

func (r *Redmine) DeleteGroup(id int) error {
	path := fmt.Sprintf("groups/%v.json", id)
	return r.delete(path)
}

func (r *Redmine) AddUserToGroup(groupId, userId int) error {
	type request struct {
		UserId int `json:"user_id"`
	}

	req := &request{userId}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("groups/%v/users.json", groupId)
	_, err = r.post(path, body)
	return err
}

func (r *Redmine) RemoveUserFromGroup(groupId, userId int) error {
	path := fmt.Sprintf("groups/%v/users/%v.json", groupId, userId)
	return r.delete(path)
}

type GroupInclude struct {
	Users       bool
	Memberships bool
}

func (g *GroupInclude) values() *url.Values {
	m := map[string]bool{
		"users":       g.Users,
		"memberships": g.Memberships,
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
