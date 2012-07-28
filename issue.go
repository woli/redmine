package redmine

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type Issue struct {
	Id             int
	Subject        string
	Parent         *Issue
	EstimatedHours float64
	SpentHours     float64
	AssignedTo     *User
	Priority       *IssuePriority
	DoneRatio      int
	Project        *Project
	Author         *User
	StartDate      time.Time
	DueDate        time.Time
	Tracker        *Tracker
	Description    string
	Status         *IssueStatus
	CustomFields   []*CustomField
	Journals       []*Journal
	Attachments    []*Attachment
	Relations      []*IssueRelation
	FixedVersion   *Version
	Category       *IssueCategory
	Changesets     []*Changeset
	Children       []*Issue
	CreatedOn      time.Time
	UpdatedOn      time.Time
}

type issueDec struct {
	Id             *int                `json:"id"`
	Subject        *string             `json:"subject"`
	Parent         *issueDec           `json:"parent"`
	EstimatedHours *float64            `json:"estimated_hours"`
	SpentHours     *float64            `json:"spent_hours"`
	AssignedTo     *userDec            `json:"assigned_to"`
	Priority       *issuePriorityDec   `json:"priority"`
	DoneRatio      *int                `json:"done_ratio"`
	Project        *projectDec         `json:"project"`
	Author         *userDec            `json:"author"`
	StartDate      *string             `json:"start_date"`
	DueDate        *string             `json:"due_date"`
	Tracker        *trackerDec         `json:"tracker"`
	Description    *string             `json:"description"`
	Status         *issueStatusDec     `json:"status"`
	CustomFields   []*customFieldDec   `json:"custom_fields"`
	Journals       []*journalDec       `json:"journals"`
	Attachments    []*attachmentDec    `json:"attachments"`
	Relations      []*issueRelationDec `json:"relations"`
	FixedVersion   *versionDec         `json:"fixed_version"`
	Category       *issueCategoryDec   `json:"category"`
	Changesets     []*changesetDec     `json:"changesets"`
	Children       []*issueDec         `json:"children"`
	CreatedOn      *string             `json:"created_on"`
	UpdatedOn      *string             `json:"updated_on"`
}

type issueEnc struct {
	Subject           string                 `json:"subject,omitempty"`
	ParentIssueId     int                    `json:"parent_issue_id,omitempty"`
	EstimatedHours    float64                `json:"estimated_hours,omitempty"`
	SpentHours        float64                `json:"spent_hours,omitempty"`
	AssignedToId      int                    `json:"assigned_to_id,omitempty"`
	PriorityId        int                    `json:"priority_id,omitempty"`
	DoneRatio         int                    `json:"done_ratio,omitempty"`
	ProjectId         int                    `json:"project_id,omitempty"`
	AuthorId          int                    `json:"author_id,omitempty"`
	StartDate         string                 `json:"start_date,omitempty"`
	DueDate           string                 `json:"due_date,omitempty"`
	TrackerId         int                    `json:"tracker_id,omitempty"`
	Description       string                 `json:"description,omitempty"`
	StatusId          int                    `json:"status_id,omitempty"`
	FixedVersionId    int                    `json:"fixed_version_id,omitempty"`
	CategoryId        int                    `json:"category_id,omitempty"`
	CustomFieldValues map[string]interface{} `json:"custom_field_values,omitempty"`
	Uploads           []*Upload              `json:"uploads,omitempty"`
}

func (i *issueDec) decode() *Issue {
	dec := &Issue{
		Id:             ptrToInt(i.Id),
		Subject:        ptrToString(i.Subject),
		EstimatedHours: ptrToFloat64(i.EstimatedHours),
		SpentHours:     ptrToFloat64(i.SpentHours),
		DoneRatio:      ptrToInt(i.DoneRatio),
		StartDate:      strToTime(DateLayout, ptrToString(i.StartDate)),
		DueDate:        strToTime(DateLayout, ptrToString(i.DueDate)),
		Description:    ptrToString(i.Description),
		CreatedOn:      strToTime(TimeLayout, ptrToString(i.CreatedOn)),
		UpdatedOn:      strToTime(TimeLayout, ptrToString(i.UpdatedOn)),
	}
	if i.Parent != nil {
		dec.Parent = i.Parent.decode()
	}
	if i.AssignedTo != nil {
		dec.AssignedTo = i.AssignedTo.decode()
	}
	if i.Priority != nil {
		dec.Priority = i.Priority.decode()
	}
	if i.Project != nil {
		dec.Project = i.Project.decode()
	}
	if i.Author != nil {
		dec.Author = i.Author.decode()
	}
	if i.Tracker != nil {
		dec.Tracker = i.Tracker.decode()
	}
	if i.Status != nil {
		dec.Status = i.Status.decode()
	}
	if i.FixedVersion != nil {
		dec.FixedVersion = i.FixedVersion.decode()
	}
	if i.Category != nil {
		dec.Category = i.Category.decode()
	}
	if i.CustomFields != nil {
		dec.CustomFields = make([]*CustomField, len(i.CustomFields))
		for j := range i.CustomFields {
			dec.CustomFields[j] = i.CustomFields[j].decode()
		}
	}
	if i.Journals != nil {
		dec.Journals = make([]*Journal, len(i.Journals))
		for j := range i.Journals {
			dec.Journals[j] = i.Journals[j].decode()
		}
	}
	if i.Attachments != nil {
		dec.Attachments = make([]*Attachment, len(i.Attachments))
		for j := range i.Attachments {
			dec.Attachments[j] = i.Attachments[j].decode()
		}
	}
	if i.Relations != nil {
		dec.Relations = make([]*IssueRelation, len(i.Relations))
		for j := range i.Relations {
			dec.Relations[j] = i.Relations[j].decode()
		}
	}
	if i.Changesets != nil {
		dec.Changesets = make([]*Changeset, len(i.Changesets))
		for j := range i.Changesets {
			dec.Changesets[j] = i.Changesets[j].decode()
		}
	}
	if i.Children != nil {
		dec.Children = make([]*Issue, len(i.Children))
		for j := range i.Children {
			dec.Children[j] = i.Children[j].decode()
		}
	}

	return dec
}

func (i *Issue) encode(uploads []*Upload) *issueEnc {
	enc := &issueEnc{
		Subject:        i.Subject,
		EstimatedHours: i.EstimatedHours,
		SpentHours:     i.SpentHours,
		DoneRatio:      i.DoneRatio,
		StartDate:      timeToStr(DateLayout, i.StartDate),
		DueDate:        timeToStr(DateLayout, i.DueDate),
		Description:    i.Description,
		Uploads:        uploads,
	}
	if i.Parent != nil {
		enc.ParentIssueId = i.Parent.Id
	}
	if i.AssignedTo != nil {
		enc.AssignedToId = i.AssignedTo.Id
	}
	if i.Priority != nil {
		enc.PriorityId = i.Priority.Id
	}
	if i.Project != nil {
		enc.ProjectId = i.Project.Id
	}
	if i.Author != nil {
		enc.AuthorId = i.Author.Id
	}
	if i.Tracker != nil {
		enc.TrackerId = i.Tracker.Id
	}
	if i.Status != nil {
		enc.StatusId = i.Status.Id
	}
	if i.FixedVersion != nil {
		enc.FixedVersionId = i.FixedVersion.Id
	}
	if i.Category != nil {
		enc.CategoryId = i.Category.Id
	}
	if i.CustomFields != nil {
		enc.CustomFieldValues = customFieldsToMap(i.CustomFields)
	}

	return enc
}

func unmarshalIssue(data []byte) (*Issue, error) {
	type response struct {
		Issue *issueDec `json:"issue"`
	}

	res := &response{}
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res.Issue.decode(), nil
}

func marshalIssue(issue *Issue, uploads []*Upload) ([]byte, error) {
	type request struct {
		Issue *issueEnc `json:"issue"`
	}

	req := &request{issue.encode(uploads)}
	return json.Marshal(req)
}

func (r *Redmine) GetIssues(v *url.Values) ([]*Issue, *Pagination, error) {
	return r.getIssues("issues.json", v)
}

func (r *Redmine) GetProjectIssues(projectId int, v *url.Values) ([]*Issue, *Pagination, error) {
	path := fmt.Sprintf("projects/%v/issues.json", projectId)
	return r.getIssues(path, v)
}

func (r *Redmine) getIssues(path string, v *url.Values) ([]*Issue, *Pagination, error) {
	data, err := r.get(path, v)
	if err != nil {
		return nil, nil, err
	}

	type response struct {
		Issues     []*issueDec `json:"issues"`
		TotalCount int         `json:"total_count"`
		Limit      int         `json:"limit"`
		Offset     int         `json:"offset"`
	}

	res := &response{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, err
	}

	issues := make([]*Issue, len(res.Issues))
	for i := range res.Issues {
		issues[i] = res.Issues[i].decode()
	}

	pagination := &Pagination{res.TotalCount, res.Limit, res.Offset}
	return issues, pagination, nil
}

func (r *Redmine) GetIssue(id int, inc *IssueInclude) (*Issue, error) {
	path := fmt.Sprintf("issues/%v.json", id)
	data, err := r.get(path, inc.values())
	if err != nil {
		return nil, err
	}

	return unmarshalIssue(data)
}

func (r *Redmine) CreateIssue(issue *Issue, uploads ...*Upload) (*Issue, error) {
	body, err := marshalIssue(issue, uploads)
	if err != nil {
		return nil, err
	}

	data, err := r.post("issues.json", body)
	if err != nil {
		return nil, err
	}

	return unmarshalIssue(data)
}

func (r *Redmine) UpdateIssue(issue *Issue, uploads ...*Upload) error {
	body, err := marshalIssue(issue, uploads)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("issues/%v.json", issue.Id)
	return r.put(path, body)
}

func (r *Redmine) DeleteIssue(id int) error {
	path := fmt.Sprintf("issues/%v.json", id)
	return r.delete(path)
}

type IssueInclude struct {
	Children    bool
	Attachments bool
	Relations   bool
	Changesets  bool
	Journals    bool
}

func (i *IssueInclude) values() *url.Values {
	m := map[string]bool{
		"children":    i.Children,
		"attachments": i.Attachments,
		"relations":   i.Relations,
		"changesets":  i.Changesets,
		"journals":    i.Journals,
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
