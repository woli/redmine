package redmine

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type TimeEntry struct {
	Id           int
	Hours        float64
	Comments     string
	SpentOn      time.Time
	Issue        *Issue
	Project      *Project
	Activity     *TimeActivity
	User         *User
	CustomFields []*CustomField
	CreatedOn    time.Time
	UpdatedOn    time.Time
}

type timeEntryDec struct {
	Id           *int              `json:"id"`
	Hours        *float64          `json:"hours"`
	Comments     *string           `json:"comments"`
	SpentOn      *string           `json:"spent_on"`
	Issue        *issueDec         `json:"issue"`
	Project      *projectDec       `json:"project"`
	Activity     *timeActivityDec  `json:"activity"`
	User         *userDec          `json:"user"`
	CustomFields []*customFieldDec `json:"custom_fields"`
	CreatedOn    *string           `json:"created_on"`
	UpdatedOn    *string           `json:"updated_on"`
}

type timeEntryEnc struct {
	ProjectId         int                    `json:"project_id,omitempty"`
	IssueId           int                    `json:"issue_id,omitempty"`
	UserId            int                    `json:"user_id,omitempty"`
	ActivityId        int                    `json:"activity_id,omitempty"`
	Hours             float64                `json:"hours,omitempty"`
	Comment           string                 `json:"comment,omitempty"`
	SpentOn           string                 `json:"spent_on,omitempty"`
	CustomFieldValues map[string]interface{} `json:"custom_field_values,omitempty"`
}

func (t *timeEntryDec) decode() *TimeEntry {
	dec := &TimeEntry{
		Id:        ptrToInt(t.Id),
		Hours:     ptrToFloat64(t.Hours),
		Comments:  ptrToString(t.Comments),
		SpentOn:   strToTime(DateLayout, ptrToString(t.SpentOn)),
		CreatedOn: strToTime(TimeLayout, ptrToString(t.CreatedOn)),
		UpdatedOn: strToTime(TimeLayout, ptrToString(t.UpdatedOn)),
	}
	if t.Issue != nil {
		dec.Issue = t.Issue.decode()
	}
	if t.Project != nil {
		dec.Project = t.Project.decode()
	}
	if t.Activity != nil {
		dec.Activity = t.Activity.decode()
	}
	if t.User != nil {
		dec.User = t.User.decode()
	}
	if t.CustomFields != nil {
		dec.CustomFields = make([]*CustomField, len(t.CustomFields))
		for i := range t.CustomFields {
			dec.CustomFields[i] = t.CustomFields[i].decode()
		}
	}

	return dec
}

func (t *TimeEntry) encode() *timeEntryEnc {
	enc := &timeEntryEnc{
		Hours:   t.Hours,
		Comment: t.Comments,
		SpentOn: timeToStr(DateLayout, t.SpentOn),
	}
	if t.Project != nil {
		enc.ProjectId = t.Project.Id
	}
	if t.Issue != nil {
		enc.IssueId = t.Issue.Id
	}
	if t.User != nil {
		enc.UserId = t.User.Id
	}
	if t.Activity != nil {
		enc.ActivityId = t.Activity.Id
	}
	if t.CustomFields != nil {
		enc.CustomFieldValues = customFieldsToMap(t.CustomFields)
	}

	return enc
}

func unmarshalTimeEntry(data []byte) (*TimeEntry, error) {
	type response struct {
		TimeEntry *timeEntryDec `json:"time_entry"`
	}

	res := &response{}
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res.TimeEntry.decode(), nil
}

func marshalTimeEntry(timeEntry *TimeEntry) ([]byte, error) {
	type request struct {
		TimeEntry *timeEntryEnc `json:"time_entry"`
	}

	req := &request{timeEntry.encode()}
	return json.Marshal(req)
}

func (r *Redmine) GetTimeEntries(v *url.Values) ([]*TimeEntry, *Pagination, error) {
	data, err := r.get("time_entries.json", v)
	if err != nil {
		return nil, nil, err
	}

	type response struct {
		TimeEntries []*timeEntryDec `json:"time_entries"`
		TotalCount  int             `json:"total_count"`
		Limit       int             `json:"limit"`
		Offset      int             `json:"offset"`
	}

	res := &response{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, err
	}

	timeEntries := make([]*TimeEntry, len(res.TimeEntries))
	for i := range res.TimeEntries {
		timeEntries[i] = res.TimeEntries[i].decode()
	}

	pagination := &Pagination{res.TotalCount, res.Limit, res.Offset}
	return timeEntries, pagination, nil
}

func (r *Redmine) GetTimeEntry(id int) (*TimeEntry, error) {
	path := fmt.Sprintf("time_entries/%v.json", id)
	data, err := r.get(path, nil)
	if err != nil {
		return nil, err
	}

	return unmarshalTimeEntry(data)
}

func (r *Redmine) CreateTimeEntry(timeEntry *TimeEntry) (*TimeEntry, error) {
	body, err := marshalTimeEntry(timeEntry)
	if err != nil {
		return nil, err
	}

	data, err := r.post("time_entries.json", body)
	if err != nil {
		return nil, err
	}

	return unmarshalTimeEntry(data)
}

func (r *Redmine) UpdateTimeEntry(timeEntry *TimeEntry) error {
	body, err := marshalTimeEntry(timeEntry)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("time_entries/%v.json", timeEntry.Id)
	return r.put(path, body)
}

func (r *Redmine) DeleteTimeEntry(id int) error {
	path := fmt.Sprintf("time_entries/%v.json", id)
	return r.delete(path)
}
