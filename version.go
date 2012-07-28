package redmine

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"
)

type VersionStatus string

const (
	VERSION_STATUS_OPEN   VersionStatus = "open"
	VERSION_STATUS_LOCKED VersionStatus = "locked"
	VERSION_STATUS_CLOSED VersionStatus = "closed"
)

type VersionSharing string

const (
	VERSION_SHARING_NONE        VersionSharing = "none"
	VERSION_SHARING_DESCENDANTS VersionSharing = "descendants"
	VERSION_SHARING_HIERARCHY   VersionSharing = "hierarchy"
	VERSION_SHARING_TREE        VersionSharing = "tree"
	VERSION_SHARING_SYSTEM      VersionSharing = "system"
)

type Version struct {
	Id           int
	Name         string
	Project      *Project
	Description  string
	Status       VersionStatus
	DueDate      time.Time
	CustomFields []*CustomField
	Sharing      VersionSharing
	CreatedOn    time.Time
	UpdatedOn    time.Time
}

type versionDec struct {
	Id           *int              `json:"id"`
	Name         *string           `json:"name"`
	Project      *projectDec       `json:"project"`
	Description  *string           `json:"description"`
	Status       *string           `json:"status"`
	DueDate      *string           `json:"due_date"`
	CustomFields []*customFieldDec `json:"custom_fields"`
	CreatedOn    *string           `json:"created_on"`
	UpdatedOn    *string           `json:"updated_on"`
}

type versionEnc struct {
	Name              string                 `json:"name,omitempty"`
	Description       string                 `json:"description,omitempty"`
	Status            string                 `json:"status,omitempty"`
	DueDate           string                 `json:"due_date,omitempty"`
	CustomFieldValues map[string]interface{} `json:"custom_field_values,omitempty"`
	Sharing           string                 `json:"sharing,omitempty"`
}

func (v *versionDec) decode() *Version {
	dec := &Version{
		Id:          ptrToInt(v.Id),
		Name:        ptrToString(v.Name),
		Description: ptrToString(v.Description),
		Status:      VersionStatus(ptrToString(v.Status)),
		DueDate:     strToTime(DateLayout, ptrToString(v.DueDate)),
		CreatedOn:   strToTime(TimeLayout, ptrToString(v.CreatedOn)),
		UpdatedOn:   strToTime(TimeLayout, ptrToString(v.UpdatedOn)),
	}
	if v.Project != nil {
		dec.Project = v.Project.decode()
	}
	if v.CustomFields != nil {
		dec.CustomFields = make([]*CustomField, len(v.CustomFields))
		for i := range v.CustomFields {
			dec.CustomFields[i] = v.CustomFields[i].decode()
		}
	}

	return dec
}

func (v *Version) encode() *versionEnc {
	enc := &versionEnc{
		Name:        v.Name,
		Description: v.Description,
		Status:      string(v.Status),
		DueDate:     timeToStr(DateLayout, v.DueDate),
		Sharing:     string(v.Sharing),
	}
	if v.CustomFields != nil {
		enc.CustomFieldValues = customFieldsToMap(v.CustomFields)
	}

	return enc
}

func unmarshalVersion(data []byte) (*Version, error) {
	type response struct {
		Version *versionDec `json:"version"`
	}

	res := &response{}
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res.Version.decode(), nil
}

func marshalVersion(version *Version) ([]byte, error) {
	type request struct {
		Version *versionEnc `json:"version"`
	}

	req := &request{version.encode()}
	return json.Marshal(req)
}

func (r *Redmine) GetVersions(projectId int, v *url.Values) ([]*Version, *Pagination, error) {
	path := fmt.Sprintf("projects/%v/versions.json", projectId)
	data, err := r.get(path, v)
	if err != nil {
		return nil, nil, err
	}

	type response struct {
		Versions   []*versionDec `json:"versions"`
		TotalCount int           `json:"total_count"`
		Limit      int           `json:"limit"`
		Offset     int           `json:"offset"`
	}

	res := &response{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, err
	}

	versions := make([]*Version, len(res.Versions))
	for i := range res.Versions {
		versions[i] = res.Versions[i].decode()
	}

	pagination := &Pagination{res.TotalCount, res.Limit, res.Offset}
	return versions, pagination, nil
}

func (r *Redmine) GetVersion(id int) (*Version, error) {
	path := fmt.Sprintf("versions/%v.json", id)
	data, err := r.get(path, nil)
	if err != nil {
		return nil, err
	}

	return unmarshalVersion(data)
}

func (r *Redmine) CreateVersion(version *Version) (*Version, error) {
	if version.Project == nil {
		return nil, errors.New("Missing required field: Project")
	}

	body, err := marshalVersion(version)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("projects/%v/versions.json", version.Project.Id)
	data, err := r.post(path, body)
	if err != nil {
		return nil, err
	}

	return unmarshalVersion(data)
}

func (r *Redmine) UpdateVersion(version *Version) error {
	body, err := marshalVersion(version)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("versions/%v.json", version.Id)
	return r.put(path, body)
}

func (r *Redmine) DeleteVersion(id int) error {
	path := fmt.Sprintf("versions/%v.json", id)
	return r.delete(path)
}
