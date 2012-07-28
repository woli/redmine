package redmine

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type User struct {
	Id           int
	Login        string
	Password     string
	FirstName    string
	LastName     string
	Name         string
	Mail         string
	CustomFields []*CustomField
	Memberships  []*ProjectMembership
	Groups       []*Group
	CreatedOn    time.Time
	LastLoginOn  time.Time
}

type userDec struct {
	Id           *int                    `json:"id"`
	Login        *string                 `json:"login"`
	FirstName    *string                 `json:"firstname"`
	LastName     *string                 `json:"lastname"`
	Name         *string                 `json:"name"`
	Mail         *string                 `json:"mail"`
	CustomFields []*customFieldDec       `json:"custom_fields"`
	Memberships  []*projectMembershipDec `json:"memberships"`
	Groups       []*groupDec             `json:"groups"`
	CreatedOn    *string                 `json:"created_on"`
	LastLoginOn  *string                 `json:"last_login_on"`
}

type userEnc struct {
	Login             string                 `json:"login,omitempty"`
	Password          string                 `json:"password,omitempty"`
	FirstName         string                 `json:"firstname,omitempty"`
	LastName          string                 `json:"lastname,omitempty"`
	Mail              string                 `json:"mail,omitempty"`
	CustomFieldValues map[string]interface{} `json:"custom_field_values,omitempty"`
}

func (u *userDec) decode() *User {
	dec := &User{
		Id:          ptrToInt(u.Id),
		Login:       ptrToString(u.Login),
		FirstName:   ptrToString(u.FirstName),
		LastName:    ptrToString(u.LastName),
		Name:        ptrToString(u.Name),
		Mail:        ptrToString(u.Mail),
		CreatedOn:   strToTime(TimeLayout, ptrToString(u.CreatedOn)),
		LastLoginOn: strToTime(TimeLayout, ptrToString(u.LastLoginOn)),
	}
	if u.CustomFields != nil {
		dec.CustomFields = make([]*CustomField, len(u.CustomFields))
		for i := range u.CustomFields {
			dec.CustomFields[i] = u.CustomFields[i].decode()
		}
	}
	if u.Memberships != nil {
		dec.Memberships = make([]*ProjectMembership, len(u.Memberships))
		for i := range u.Memberships {
			dec.Memberships[i] = u.Memberships[i].decode()
		}
	}
	if u.Groups != nil {
		dec.Groups = make([]*Group, len(u.Groups))
		for i := range u.Groups {
			dec.Groups[i] = u.Groups[i].decode()
		}
	}

	return dec
}

func (u *User) encode() *userEnc {
	enc := &userEnc{
		Login:     u.Login,
		Password:  u.Password,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Mail:      u.Mail,
	}
	if u.CustomFields != nil {
		enc.CustomFieldValues = customFieldsToMap(u.CustomFields)
	}

	return enc
}

func unmarshalUser(data []byte) (*User, error) {
	type response struct {
		User *userDec `json:"user"`
	}

	res := &response{}
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res.User.decode(), nil
}

func marshalUser(user *User) ([]byte, error) {
	type request struct {
		User *userEnc `json:"user"`
	}

	req := &request{user.encode()}
	return json.Marshal(req)
}

func (r *Redmine) GetUsers(v *url.Values) ([]*User, *Pagination, error) {
	data, err := r.get("users.json", v)
	if err != nil {
		return nil, nil, err
	}

	type response struct {
		Users      []*userDec `json:"users"`
		TotalCount int        `json:"total_count"`
		Limit      int        `json:"limit"`
		Offset     int        `json:"offset"`
	}

	res := &response{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, err
	}

	users := make([]*User, len(res.Users))
	for i := range res.Users {
		users[i] = res.Users[i].decode()
	}

	pagination := &Pagination{res.TotalCount, res.Limit, res.Offset}
	return users, pagination, nil
}

func (r *Redmine) GetUser(id int, inc *UserInclude) (*User, error) {
	path := fmt.Sprintf("users/%v.json", id)
	data, err := r.get(path, inc.values())
	if err != nil {
		return nil, err
	}

	return unmarshalUser(data)
}

func (r *Redmine) CreateUser(user *User) (*User, error) {
	body, err := marshalUser(user)
	if err != nil {
		return nil, err
	}

	data, err := r.post("users.json", body)
	if err != nil {
		return nil, err
	}

	return unmarshalUser(data)
}

func (r *Redmine) UpdateUser(user *User) error {
	body, err := marshalUser(user)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("users/%v.json", user.Id)
	return r.put(path, body)
}

func (r *Redmine) DeleteUser(id int) error {
	path := fmt.Sprintf("users/%v.json", id)
	return r.delete(path)
}

type UserInclude struct {
	Memberships bool
	Groups      bool
}

func (u *UserInclude) values() *url.Values {
	m := map[string]bool{
		"memberships": u.Memberships,
		"groups":      u.Groups,
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
