// Package redmine provides access to the Redmine API.
//
// See http://www.redmine.org/projects/redmine/wiki/Rest_api
//
// Redmine Version: 2.2.3
package redmine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Service struct {
	baseUrl         string
	client          *http.Client
	auth            Authenticator
	switchUser      string
	Uploads         *UploadsService
	Issues          *IssuesService
	Projects        *ProjectsService
	Memberships     *MembershipsService
	Users           *UsersService
	TimeEntries     *TimeEntriesService
	News            *NewsService
	Relations       *RelationsService
	Versions        *VersionsService
	Wiki            *WikiService
	Queries         *QueriesService
	Attachments     *AttachmentsService
	IssueStatuses   *IssueStatusesService
	Trackers        *TrackersService
	Enumerations    *Enumerations
	IssueCategories *IssueCategoriesService
	Roles           *RolesService
	Groups          *GroupsService
}

type service struct {
	s *Service
}

type UploadsService service
type IssuesService service
type ProjectsService service
type MembershipsService service
type UsersService service
type TimeEntriesService service
type NewsService service
type RelationsService service
type VersionsService service
type WikiService service
type QueriesService service
type AttachmentsService service
type IssueStatusesService service
type TrackersService service
type Enumerations struct {
	DocumentCategories  *DocumentCategoriesService
	IssuePriorities     *IssuePrioritiesService
	TimeEntryActivities *TimeEntryActivitiesService
}
type DocumentCategoriesService service
type IssuePrioritiesService service
type TimeEntryActivitiesService service
type IssueCategoriesService service
type RolesService service
type GroupsService service

func New(baseUrl string, auth Authenticator, client *http.Client) (*Service, error) {
	if auth == nil {
		return nil, errors.New("auth is nil")
	}
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{
		baseUrl: baseUrl,
		auth:    auth,
		client:  client,
	}
	s.Uploads = &UploadsService{s}
	s.Issues = &IssuesService{s}
	s.Projects = &ProjectsService{s}
	s.Memberships = &MembershipsService{s}
	s.Users = &UsersService{s}
	s.TimeEntries = &TimeEntriesService{s}
	s.News = &NewsService{s}
	s.Relations = &RelationsService{s}
	s.Versions = &VersionsService{s}
	s.Wiki = &WikiService{s}
	s.Queries = &QueriesService{s}
	s.Attachments = &AttachmentsService{s}
	s.IssueStatuses = &IssueStatusesService{s}
	s.Trackers = &TrackersService{s}
	s.Enumerations = &Enumerations{
		DocumentCategories:  &DocumentCategoriesService{s},
		IssuePriorities:     &IssuePrioritiesService{s},
		TimeEntryActivities: &TimeEntryActivitiesService{s},
	}
	s.IssueCategories = &IssueCategoriesService{s}
	s.Roles = &RolesService{s}
	s.Groups = &GroupsService{s}
	return s, nil
}

func (s *Service) SwitchUser(username string) {
	s.switchUser = username
}

//-------------------------------------------------------------------------
// autentication
//-------------------------------------------------------------------------

type Authenticator interface {
	SetAuth(req *http.Request)
}

type BasicAuth struct {
	Username string
	Password string
}

func (a *BasicAuth) SetAuth(req *http.Request) {
	req.SetBasicAuth(a.Username, a.Password)
}

type ApiKeyAuth struct {
	ApiKey string
}

func (a *ApiKeyAuth) SetAuth(req *http.Request) {
	req.SetBasicAuth(a.ApiKey, "")
}

//-------------------------------------------------------------------------
// common
//-------------------------------------------------------------------------

func resolveRelative(basestr, relstr string) string {
	u, _ := url.Parse(basestr)
	rel, _ := url.Parse(relstr)
	u = u.ResolveReference(rel)
	us := u.String()
	us = strings.Replace(us, "%7B", "{", -1)
	us = strings.Replace(us, "%7D", "}", -1)
	return us
}

func (s *Service) doRequest(method, urlStr string, body io.Reader) ([]byte, error) {
	req, _ := http.NewRequest(method, urlStr, body)
	s.auth.SetAuth(req)
	req.Header.Set("Content-Type", "application/json")
	if s.switchUser != "" {
		req.Header.Set("X-Redmine-Switch-User", s.switchUser)
	}
	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = checkResponse(res.StatusCode, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func checkResponse(statusCode int, data []byte) error {
	if statusCode >= 200 && statusCode <= 299 {
		return nil
	}
	if statusCode == 422 {
		r := struct {
			Errors []string `json:"errors"`
		}{}
		err := json.Unmarshal(data, &r)
		if err != nil {
			return err
		}
		if len(r.Errors) > 0 {
			return fmt.Errorf("%v", strings.Join(r.Errors, "; "))
		}
	}
	return fmt.Errorf("%v", statusCode)
}

type Id struct {
	Id int `json:"id"`
}

type Name struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

//-------------------------------------------------------------------------
// custom fields
//-------------------------------------------------------------------------

type CustomField struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type customField struct {
	Id    int    `json:"id,omitempty"`
	Value string `json:"value,omitempty"`
}

//-------------------------------------------------------------------------
// uploads
//-------------------------------------------------------------------------

type Upload struct {
	Token       string `json:"token,omitempty"`
	Filename    string `json:"filename,omitempty"`
	Description string `json:"description,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

//-------------------------------------------------------------------------
// upload file
//-------------------------------------------------------------------------

type UploadsUploadCall struct {
	s    *Service
	data []byte
}

func (r *UploadsService) Upload(data []byte) *UploadsUploadCall {
	return &UploadsUploadCall{r.s, data}
}

func (c *UploadsUploadCall) Do() (string, error) {
	body := new(bytes.Buffer)
	body.Write(c.data)
	urlStr := resolveRelative(c.s.baseUrl, "uploads.json")
	req, _ := http.NewRequest("POST", urlStr, body)
	c.s.auth.SetAuth(req)
	req.Header.Set("Content-Type", "application/octet-stream")
	if c.s.switchUser != "" {
		req.Header.Set("X-Redmine-Switch-User", c.s.switchUser)
	}
	res, err := c.s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	err = checkResponse(res.StatusCode, data)
	if err != nil {
		return "", err
	}
	ret := struct {
		Upload struct {
			Token string `json:"token"`
		} `json:"upload"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return "", err
	}
	return ret.Upload.Token, nil
}

//-------------------------------------------------------------------------
// issues
//-------------------------------------------------------------------------

type IssueFeed struct {
	Issues     []*Issue `json:"issues"`
	TotalCount int      `json:"total_count"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
}

type Issue struct {
	Id             int               `json:"id"`
	DoneRatio      int               `json:"done_ratio"`
	Priority       *Name             `json:"priority"`
	Status         *Name             `json:"status"`
	Subject        string            `json:"subject"`
	FixedVersion   *Name             `json:"fixed_version"`
	UpdatedOn      string            `json:"updated_on"`
	Project        *Name             `json:"project"`
	Tracker        *Name             `json:"tracker"`
	Author         *Name             `json:"author"`
	CreatedOn      string            `json:"created_on"`
	StartDate      string            `json:"start_date"`
	DueDate        string            `json:"due_date"`
	SpentHours     float64           `json:"spent_hours"`
	EstimatedHours float64           `json:"estimated_hours"`
	Parent         *Id               `json:"parent"`
	AssignedTo     *Name             `json:"assigned_to"`
	Description    string            `json:"description"`
	Category       *Name             `json:"category"`
	CustomFields   []*CustomField    `json:"custom_fields"`
	Relations      []*IssueRelation  `json:"relations"`
	Children       []*IssueChild     `json:"children"`
	Attachments    []*Attachment     `json:"attachments"`
	Journals       []*IssueJournal   `json:"journals"`
	Changesets     []*IssueChangeset `json:"changesets"`
	// TODO: watchers (since 2.3)
}

type IssueRelation struct {
	Id           int    `json:"id"`
	IssueId      int    `json:"issue_id"`
	IssueToId    int    `json:"issue_to_id"`
	RelationType string `json:"relation_type"`
	Delay        Nint   `json:"delay"`
}

type IssueChild struct {
	Id      int    `json:"id"`
	Subject string `json:"subject"`
	Tracker *Name  `json:"tracker"`
}

type IssueJournalDetail struct {
	Name     string `json:"name"`
	Property string `json:"property"`
	NewValue string `json:"new_value"`
}

type IssueJournal struct {
	Id        int                   `json:"id"`
	User      *Name                 `json:"user"`
	Details   []*IssueJournalDetail `json:"details"`
	Notes     string                `json:"notes"`
	CreatedOn string                `json:"created_on"`
}

type IssueChangeset struct {
	Comments    string `json:"comments"`
	Revision    string `json:"revision"`
	CommittedOn string `json:"committed_on"`
	User        *Name  `json:"user"`
}

type Nint int

func (n *Nint) UnmarshalJSON(b []byte) (err error) {
	if string(b) == "null" {
		return nil
	}
	return json.Unmarshal(b, (*int)(n))
}

type issue struct {
	Subject        string         `json:"subject,omitempty"`
	ParentIssueId  int            `json:"parent_issue_id,omitempty"`
	EstimatedHours float64        `json:"estimated_hours,omitempty"`
	SpentHours     float64        `json:"spent_hours,omitempty"`
	AssignedToId   int            `json:"assigned_to_id,omitempty"`
	PriorityId     int            `json:"priority_id,omitempty"`
	DoneRatio      int            `json:"done_ratio,omitempty"`
	ProjectId      int            `json:"project_id,omitempty"`
	AuthorId       int            `json:"author_id,omitempty"`
	StartDate      string         `json:"start_date,omitempty"`
	DueDate        string         `json:"due_date,omitempty"`
	TrackerId      int            `json:"tracker_id,omitempty"`
	Description    string         `json:"description,omitempty"`
	StatusId       int            `json:"status_id,omitempty"`
	FixedVersionId int            `json:"fixed_version_id,omitempty"`
	CategoryId     int            `json:"category_id,omitempty"`
	CustomFields   []*customField `json:"custom_fields,omitempty"`
	Uploads        []*Upload      `json:"uploads,omitempty"`
}

func (r *Issue) toSend(uploads []*Upload) *issue {
	newIssue := new(issue)
	newIssue.Subject = r.Subject
	if r.Parent != nil {
		newIssue.ParentIssueId = r.Parent.Id
	}
	newIssue.EstimatedHours = r.EstimatedHours
	newIssue.SpentHours = r.SpentHours
	if r.AssignedTo != nil {
		newIssue.AssignedToId = r.AssignedTo.Id
	}
	if r.Priority != nil {
		newIssue.PriorityId = r.Priority.Id
	}
	newIssue.DoneRatio = r.DoneRatio
	if r.Project != nil {
		newIssue.ProjectId = r.Project.Id
	}
	if r.AssignedTo != nil {
		newIssue.AuthorId = r.Author.Id
	}
	newIssue.StartDate = r.StartDate
	newIssue.DueDate = r.DueDate
	if r.Tracker != nil {
		newIssue.TrackerId = r.Tracker.Id
	}
	newIssue.Description = r.Description
	if r.Status != nil {
		newIssue.StatusId = r.Status.Id
	}
	if r.FixedVersion != nil {
		newIssue.FixedVersionId = r.FixedVersion.Id
	}
	if r.Category != nil {
		newIssue.CategoryId = r.Category.Id
	}
	if r.CustomFields != nil {
		newIssue.CustomFields = make([]*customField, len(r.CustomFields))
		for i, cf := range r.CustomFields {
			newIssue.CustomFields[i] = &customField{
				Id:    cf.Id,
				Value: cf.Value,
			}
		}
	}
	newIssue.Uploads = uploads
	return newIssue
}

//-------------------------------------------------------------------------
// list issues
//-------------------------------------------------------------------------

type IssuesListCall struct {
	s       *Service
	options map[string]interface{}
	filters map[string]string
}

func (r *IssuesService) List() *IssuesListCall {
	return &IssuesListCall{
		s:       r.s,
		options: make(map[string]interface{}),
		filters: make(map[string]string),
	}
}

func (c *IssuesListCall) Offset(offset int) *IssuesListCall {
	c.options["offset"] = offset
	return c
}

func (c *IssuesListCall) Limit(limit int) *IssuesListCall {
	c.options["limit"] = limit
	return c
}

func (c *IssuesListCall) Sort(sort string) *IssuesListCall {
	c.options["sort"] = sort
	return c
}

func (c *IssuesListCall) Filter(key, value string) *IssuesListCall {
	c.options[key] = value
	return c
}

func (c *IssuesListCall) Do() (*IssueFeed, error) {
	params := make(url.Values)
	for _, opt := range []string{"offset", "limit", "sort"} {
		if v, ok := c.options[opt]; ok {
			params.Set(opt, fmt.Sprintf("%v", v))
		}
	}
	for k, v := range c.filters {
		params.Set(k, v)
	}
	urlStr := resolveRelative(c.s.baseUrl, "issues.json")
	urlStr += "?" + params.Encode()
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := new(IssueFeed)
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

//-------------------------------------------------------------------------
// get issue
//-------------------------------------------------------------------------

type IssuesGetCall struct {
	s       *Service
	issueId int
	options map[string]interface{}
}

func (r *IssuesService) Get(issueId int) *IssuesGetCall {
	return &IssuesGetCall{
		s:       r.s,
		issueId: issueId,
		options: make(map[string]interface{}),
	}
}

func (c *IssuesGetCall) Children(children bool) *IssuesGetCall {
	c.options["children"] = children
	return c
}

func (c *IssuesGetCall) Attachments(attachments bool) *IssuesGetCall {
	c.options["attachments"] = attachments
	return c
}

func (c *IssuesGetCall) Relations(relations bool) *IssuesGetCall {
	c.options["relations"] = relations
	return c
}

func (c *IssuesGetCall) Changesets(changesets bool) *IssuesGetCall {
	c.options["changesets"] = changesets
	return c
}

func (c *IssuesGetCall) Journals(journals bool) *IssuesGetCall {
	c.options["journals"] = journals
	return c
}

/*
func (c *IssuesGetCall) Watchers(watchers bool) *IssuesGetCall {
	c.options["watchers"] = watchers
	return c
}
*/

func (c *IssuesGetCall) Do() (*Issue, error) {
	params := make(url.Values)
	var include []string
	for _, inc := range []string{"children", "attachments", "relations", "changesets", "journals", "watchers"} {
		v, ok := c.options[inc]
		if !ok {
			continue
		}
		if b, ok := v.(bool); ok && b {
			include = append(include, inc)
		}
	}
	if len(include) > 0 {
		params.Set("include", strings.Join(include, ","))
	}
	urlStr := resolveRelative(c.s.baseUrl, "issues/{issueId}.json")
	urlStr = strings.Replace(urlStr, "{issueId}", strconv.Itoa(c.issueId), 1)
	urlStr += "?" + params.Encode()
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Issue *Issue `json:"issue"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Issue, nil
}

//-------------------------------------------------------------------------
// insert issue
//-------------------------------------------------------------------------

type IssuesInsertCall struct {
	s       *Service
	issue   *Issue
	uploads []*Upload
}

func (r *IssuesService) Insert(issue *Issue, uploads ...*Upload) *IssuesInsertCall {
	return &IssuesInsertCall{r.s, issue, uploads}
}

func (c *IssuesInsertCall) Do() (*Issue, error) {
	v := struct {
		Issue *issue `json:"issue,omitempty"`
	}{
		c.issue.toSend(c.uploads),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return nil, err
	}
	urlStr := resolveRelative(c.s.baseUrl, "issues.json")
	data, err := c.s.doRequest("POST", urlStr, body)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Issue *Issue `json:"issue"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Issue, nil
}

//-------------------------------------------------------------------------
// update issue
//-------------------------------------------------------------------------

type IssuesUpdateCall struct {
	s       *Service
	issue   *Issue
	uploads []*Upload
}

func (r *IssuesService) Update(issue *Issue, uploads ...*Upload) *IssuesUpdateCall {
	return &IssuesUpdateCall{r.s, issue, uploads}
}

func (c *IssuesUpdateCall) Do() error {
	v := struct {
		Issue *issue `json:"issue,omitempty"`
	}{
		c.issue.toSend(c.uploads),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return err
	}
	urlStr := resolveRelative(c.s.baseUrl, "issues/{issueId}.json")
	urlStr = strings.Replace(urlStr, "{issueId}", strconv.Itoa(c.issue.Id), 1)
	_, err = c.s.doRequest("PUT", urlStr, body)
	return err
}

//-------------------------------------------------------------------------
// delete issue
//-------------------------------------------------------------------------

type IssuesDeleteCall struct {
	s       *Service
	issueId int
}

func (r *IssuesService) Delete(issueId int) *IssuesDeleteCall {
	return &IssuesDeleteCall{r.s, issueId}
}

func (c *IssuesDeleteCall) Do() error {
	urlStr := resolveRelative(c.s.baseUrl, "issues/{issueId}.json")
	urlStr = strings.Replace(urlStr, "{issueId}", strconv.Itoa(c.issueId), 1)
	_, err := c.s.doRequest("DELETE", urlStr, nil)
	return err
}

//-------------------------------------------------------------------------
// projects
//-------------------------------------------------------------------------

type ProjectFeed struct {
	Projects   []*Project `json:"projects"`
	TotalCount int        `json:"total_count"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

type Project struct {
	Id              int            `json:"id"`
	Identifier      string         `json:"identifier"`
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	Homepage        string         `json:"homepage"`
	Parent          *Name          `json:"parent"`
	Trackers        []*Name        `json:"trackers"`
	IssueCategories []*Name        `json:"issue_categories"`
	CustomFields    []*CustomField `json:"custom_fields"`
	CreatedOn       string         `json:"created_on"`
	UpdatedOn       string         `json:"updated_on"`
}

type project struct {
	Identifier   string         `json:"identifier,omitempty"`
	Name         string         `json:"name,omitempty"`
	Description  string         `json:"description,omitempty"`
	Homepage     string         `json:"homepage,omitempty"`
	ParentId     int            `json:"parent_id,omitempty"`
	CustomFields []*customField `json:"custom_fields,omitempty"`
}

func (r *Project) toSend() *project {
	newProject := new(project)
	newProject.Identifier = r.Identifier
	newProject.Name = r.Name
	newProject.Description = r.Description
	newProject.Homepage = r.Homepage
	if r.Parent != nil {
		newProject.ParentId = r.Parent.Id
	}
	if r.CustomFields != nil {
		newProject.CustomFields = make([]*customField, len(r.CustomFields))
		for i, cf := range r.CustomFields {
			newProject.CustomFields[i] = &customField{
				Id:    cf.Id,
				Value: cf.Value,
			}
		}
	}
	return newProject
}

//-------------------------------------------------------------------------
// list projects
//-------------------------------------------------------------------------

type ProjectsListCall struct {
	s       *Service
	options map[string]interface{}
}

func (r *ProjectsService) List() *ProjectsListCall {
	return &ProjectsListCall{
		s:       r.s,
		options: make(map[string]interface{}),
	}
}

func (c *ProjectsListCall) Offset(offset int) *ProjectsListCall {
	c.options["offset"] = offset
	return c
}

func (c *ProjectsListCall) Limit(limit int) *ProjectsListCall {
	c.options["limit"] = limit
	return c
}

func (c *ProjectsListCall) Do() (*ProjectFeed, error) {
	params := make(url.Values)
	for _, opt := range []string{"offset", "limit"} {
		if v, ok := c.options[opt]; ok {
			params.Set(opt, fmt.Sprintf("%v", v))
		}
	}
	urlStr := resolveRelative(c.s.baseUrl, "projects.json")
	urlStr += "?" + params.Encode()
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := new(ProjectFeed)
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

//-------------------------------------------------------------------------
// get project
//-------------------------------------------------------------------------

type ProjectsGetCall struct {
	s         *Service
	projectId int
	options   map[string]interface{}
}

func (r *ProjectsService) Get(projectId int) *ProjectsGetCall {
	return &ProjectsGetCall{
		s:         r.s,
		projectId: projectId,
		options:   make(map[string]interface{}),
	}
}

func (c *ProjectsGetCall) Trackers(trackers bool) *ProjectsGetCall {
	c.options["trackers"] = trackers
	return c
}

func (c *ProjectsGetCall) IssueCategories(issueCategories bool) *ProjectsGetCall {
	c.options["issue_categories"] = issueCategories
	return c
}

func (c *ProjectsGetCall) Do() (*Project, error) {
	params := make(url.Values)
	var include []string
	for _, inc := range []string{"trackers", "issue_categories"} {
		v, ok := c.options[inc]
		if !ok {
			continue
		}
		if b, ok := v.(bool); ok && b {
			include = append(include, inc)
		}
	}
	if len(include) > 0 {
		params.Set("include", strings.Join(include, ","))
	}
	urlStr := resolveRelative(c.s.baseUrl, "projects/{projectId}.json")
	urlStr = strings.Replace(urlStr, "{projectId}", strconv.Itoa(c.projectId), 1)
	urlStr += "?" + params.Encode()
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Project *Project `json:"project"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Project, nil
}

//-------------------------------------------------------------------------
// insert project
//-------------------------------------------------------------------------

type ProjectsInsertCall struct {
	s       *Service
	project *Project
}

func (r *ProjectsService) Insert(project *Project) *ProjectsInsertCall {
	return &ProjectsInsertCall{r.s, project}
}

func (c *ProjectsInsertCall) Do() (*Project, error) {
	v := struct {
		Project *project `json:"project,omitempty"`
	}{
		c.project.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return nil, err
	}
	urlStr := resolveRelative(c.s.baseUrl, "projects.json")
	data, err := c.s.doRequest("POST", urlStr, body)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Project *Project `json:"project"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Project, nil
}

//-------------------------------------------------------------------------
// update project
//-------------------------------------------------------------------------

type ProjectsUpdateCall struct {
	s       *Service
	project *Project
}

func (r *ProjectsService) Update(project *Project) *ProjectsUpdateCall {
	return &ProjectsUpdateCall{r.s, project}
}

func (c *ProjectsUpdateCall) Do() error {
	v := struct {
		Project *project `json:"project,omitempty"`
	}{
		c.project.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return err
	}
	urlStr := resolveRelative(c.s.baseUrl, "projects/{projectId}.json")
	urlStr = strings.Replace(urlStr, "{projectId}", strconv.Itoa(c.project.Id), 1)
	_, err = c.s.doRequest("PUT", urlStr, body)
	return err
}

//-------------------------------------------------------------------------
// delete project
//-------------------------------------------------------------------------

type ProjectsDeleteCall struct {
	s         *Service
	projectId int
}

func (r *ProjectsService) Delete(projectId int) *ProjectsDeleteCall {
	return &ProjectsDeleteCall{r.s, projectId}
}

func (c *ProjectsDeleteCall) Do() error {
	urlStr := resolveRelative(c.s.baseUrl, "projects/{projectId}.json")
	urlStr = strings.Replace(urlStr, "{projectId}", strconv.Itoa(c.projectId), 1)
	_, err := c.s.doRequest("DELETE", urlStr, nil)
	return err
}

//-------------------------------------------------------------------------
// memberships
//-------------------------------------------------------------------------

type MembershipFeed struct {
	Memberships []*Membership `json:"memberships"`
	TotalCount  int           `json:"total_count"`
	Limit       int           `json:"limit"`
	Offset      int           `json:"offset"`
}

type Membership struct {
	Id      int               `json:"id"`
	Project *Name             `json:"project"`
	User    *Name             `json:"user"`
	Group   *Name             `json:"group"`
	Roles   []*MembershipRole `json:"roles"`
}

type MembershipRole struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Inherited bool   `json:"inherited"`
}

type membership struct {
	UserId  int   `json:"user_id,omitempty"`
	RoleIds []int `json:"role_ids,omitempty"`
}

func (r *Membership) toSend() *membership {
	newMembership := new(membership)
	if r.User != nil {
		newMembership.UserId = r.User.Id
	}
	if r.Roles != nil {
		newMembership.RoleIds = make([]int, len(r.Roles))
		for i, role := range r.Roles {
			newMembership.RoleIds[i] = role.Id
		}
	}
	return newMembership
}

//-------------------------------------------------------------------------
// list memberships
//-------------------------------------------------------------------------

type MembershipsListCall struct {
	s         *Service
	projectId int
	options   map[string]interface{}
}

func (r *MembershipsService) List(projectId int) *MembershipsListCall {
	return &MembershipsListCall{
		s:         r.s,
		projectId: projectId,
		options:   make(map[string]interface{}),
	}
}

func (c *MembershipsListCall) Offset(offset int) *MembershipsListCall {
	c.options["offset"] = offset
	return c
}

func (c *MembershipsListCall) Limit(limit int) *MembershipsListCall {
	c.options["limit"] = limit
	return c
}

func (c *MembershipsListCall) Do() (*MembershipFeed, error) {
	params := make(url.Values)
	for _, opt := range []string{"offset", "limit"} {
		if v, ok := c.options[opt]; ok {
			params.Set(opt, fmt.Sprintf("%v", v))
		}
	}
	urlStr := resolveRelative(c.s.baseUrl, "projects/{projectId}/memberships.json")
	urlStr = strings.Replace(urlStr, "{projectId}", strconv.Itoa(c.projectId), 1)
	urlStr += "?" + params.Encode()
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := new(MembershipFeed)
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

//-------------------------------------------------------------------------
// get membership
//-------------------------------------------------------------------------

type MembershipsGetCall struct {
	s            *Service
	membershipId int
}

func (r *MembershipsService) Get(membershipId int) *MembershipsGetCall {
	return &MembershipsGetCall{r.s, membershipId}
}

func (c *MembershipsGetCall) Do() (*Membership, error) {
	urlStr := resolveRelative(c.s.baseUrl, "memberships/{membershipId}.json")
	urlStr = strings.Replace(urlStr, "{membershipId}", strconv.Itoa(c.membershipId), 1)
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Membership *Membership `json:"membership"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Membership, nil
}

//-------------------------------------------------------------------------
// insert membership
//-------------------------------------------------------------------------

type MembershipsInsertCall struct {
	s          *Service
	membership *Membership
}

func (r *MembershipsService) Insert(membership *Membership) *MembershipsInsertCall {
	return &MembershipsInsertCall{r.s, membership}
}

func (c *MembershipsInsertCall) Do() (*Membership, error) {
	v := struct {
		Membership *membership `json:"membership,omitempty"`
	}{
		c.membership.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return nil, err
	}
	urlStr := resolveRelative(c.s.baseUrl, "projects/{projectId}/memberships.json")
	urlStr = strings.Replace(urlStr, "{projectId}", strconv.Itoa(c.membership.Project.Id), 1)
	data, err := c.s.doRequest("POST", urlStr, body)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Membership *Membership `json:"membership"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Membership, nil
}

//-------------------------------------------------------------------------
// update membership
//-------------------------------------------------------------------------

type MembershipsUpdateCall struct {
	s          *Service
	membership *Membership
}

func (r *MembershipsService) Update(membership *Membership) *MembershipsUpdateCall {
	return &MembershipsUpdateCall{r.s, membership}
}

func (c *MembershipsUpdateCall) Do() error {
	v := struct {
		Membership *membership `json:"membership,omitempty"`
	}{
		c.membership.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return err
	}
	urlStr := resolveRelative(c.s.baseUrl, "memberships/{membershipId}.json")
	urlStr = strings.Replace(urlStr, "{membershipId}", strconv.Itoa(c.membership.Id), 1)
	_, err = c.s.doRequest("PUT", urlStr, body)
	return err
}

//-------------------------------------------------------------------------
// delete membership
//-------------------------------------------------------------------------

type MembershipsDeleteCall struct {
	s            *Service
	membershipId int
}

func (r *MembershipsService) Delete(membershipId int) *MembershipsDeleteCall {
	return &MembershipsDeleteCall{r.s, membershipId}
}

func (c *MembershipsDeleteCall) Do() error {
	urlStr := resolveRelative(c.s.baseUrl, "memberships/{membershipId}.json")
	urlStr = strings.Replace(urlStr, "{membershipId}", strconv.Itoa(c.membershipId), 1)
	_, err := c.s.doRequest("DELETE", urlStr, nil)
	return err
}

//-------------------------------------------------------------------------
// users
//-------------------------------------------------------------------------

type UserFeed struct {
	Users      []*User `json:"users"`
	TotalCount int     `json:"total_count"`
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
}

type User struct {
	Id           int               `json:"id"`
	Login        string            `json:"login"`
	Password     string            `json:"password"`
	Firstname    string            `json:"firstname"`
	Lastname     string            `json:"lastname"`
	Name         string            `json:"name"`
	Mail         string            `json:"mail"`
	Memberships  []*UserMembership `json:"memberships"`
	Groups       []*Name           `json:"groups"`
	CustomFields []*CustomField    `json:"custom_fields"`
	CreatedOn    string            `json:"created_on"`
	LastLoginOn  string            `json:"last_login_on"`
}

type UserMembership struct {
	Project *Name   `json:"project"`
	Roles   []*Name `json:"roles"`
}

type user struct {
	Login        string         `json:"login,omitempty"`
	Password     string         `json:"password,omitempty"`
	Firstname    string         `json:"firstname,omitempty"`
	Lastname     string         `json:"lastname,omitempty"`
	Mail         string         `json:"mail,omitempty"`
	CustomFields []*customField `json:"custom_fields,omitempty"`
}

func (r *User) toSend() *user {
	newUser := new(user)
	newUser.Login = r.Login
	newUser.Password = r.Password
	newUser.Firstname = r.Firstname
	newUser.Lastname = r.Lastname
	newUser.Mail = r.Mail
	if r.CustomFields != nil {
		newUser.CustomFields = make([]*customField, len(r.CustomFields))
		for i, cf := range r.CustomFields {
			newUser.CustomFields[i] = &customField{
				Id:    cf.Id,
				Value: cf.Value,
			}
		}
	}
	return newUser
}

//-------------------------------------------------------------------------
// list users
//-------------------------------------------------------------------------

type UsersListCall struct {
	s       *Service
	options map[string]interface{}
}

func (r *UsersService) List() *UsersListCall {
	return &UsersListCall{
		s:       r.s,
		options: make(map[string]interface{}),
	}
}

func (c *UsersListCall) Offset(offset int) *UsersListCall {
	c.options["offset"] = offset
	return c
}

func (c *UsersListCall) Limit(limit int) *UsersListCall {
	c.options["limit"] = limit
	return c
}

func (c *UsersListCall) Do() (*UserFeed, error) {
	params := make(url.Values)
	for _, opt := range []string{"offset", "limit"} {
		if v, ok := c.options[opt]; ok {
			params.Set(opt, fmt.Sprintf("%v", v))
		}
	}
	urlStr := resolveRelative(c.s.baseUrl, "users.json")
	urlStr += "?" + params.Encode()
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := new(UserFeed)
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

//-------------------------------------------------------------------------
// get user
//-------------------------------------------------------------------------

type UsersGetCall struct {
	s       *Service
	userId  int
	options map[string]interface{}
}

func (r *UsersService) Get(userId int) *UsersGetCall {
	return &UsersGetCall{
		s:       r.s,
		userId:  userId,
		options: make(map[string]interface{}),
	}
}

func (c *UsersGetCall) Memberships(memberships bool) *UsersGetCall {
	c.options["memberships"] = memberships
	return c
}

func (c *UsersGetCall) Groups(groups bool) *UsersGetCall {
	c.options["groups"] = groups
	return c
}

func (c *UsersGetCall) Do() (*User, error) {
	params := make(url.Values)
	var include []string
	for _, inc := range []string{"memberships", "groups"} {
		v, ok := c.options[inc]
		if !ok {
			continue
		}
		if b, ok := v.(bool); ok && b {
			include = append(include, inc)
		}
	}
	if len(include) > 0 {
		params.Set("include", strings.Join(include, ","))
	}
	urlStr := resolveRelative(c.s.baseUrl, "users/{userId}.json")
	urlStr = strings.Replace(urlStr, "{userId}", strconv.Itoa(c.userId), 1)
	urlStr += "?" + params.Encode()
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		User *User `json:"user"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.User, nil
}

//-------------------------------------------------------------------------
// insert user
//-------------------------------------------------------------------------

type UsersInsertCall struct {
	s    *Service
	user *User
}

func (r *UsersService) Insert(user *User) *UsersInsertCall {
	return &UsersInsertCall{r.s, user}
}

func (c *UsersInsertCall) Do() (*User, error) {
	v := struct {
		User *user `json:"user,omitempty"`
	}{
		c.user.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return nil, err
	}
	urlStr := resolveRelative(c.s.baseUrl, "users.json")
	data, err := c.s.doRequest("POST", urlStr, body)
	if err != nil {
		return nil, err
	}
	ret := struct {
		User *User `json:"user"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.User, nil
}

//-------------------------------------------------------------------------
// update user
//-------------------------------------------------------------------------

type UsersUpdateCall struct {
	s    *Service
	user *User
}

func (r *UsersService) Update(user *User) *UsersUpdateCall {
	return &UsersUpdateCall{r.s, user}
}

func (c *UsersUpdateCall) Do() error {
	v := struct {
		User *user `json:"user,omitempty"`
	}{
		c.user.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return err
	}
	urlStr := resolveRelative(c.s.baseUrl, "users/{userId}.json")
	urlStr = strings.Replace(urlStr, "{userId}", strconv.Itoa(c.user.Id), 1)
	_, err = c.s.doRequest("PUT", urlStr, body)
	return err
}

//-------------------------------------------------------------------------
// delete user
//-------------------------------------------------------------------------

type UsersDeleteCall struct {
	s      *Service
	userId int
}

func (r *UsersService) Delete(userId int) *UsersDeleteCall {
	return &UsersDeleteCall{r.s, userId}
}

func (c *UsersDeleteCall) Do() error {
	urlStr := resolveRelative(c.s.baseUrl, "users/{userId}.json")
	urlStr = strings.Replace(urlStr, "{userId}", strconv.Itoa(c.userId), 1)
	_, err := c.s.doRequest("DELETE", urlStr, nil)
	return err
}

//-------------------------------------------------------------------------
// time entries
//-------------------------------------------------------------------------

type TimeEntryFeed struct {
	TimeEntries []*TimeEntry `json:"time_entries"`
	TotalCount  int          `json:"total_count"`
	Limit       int          `json:"limit"`
	Offset      int          `json:"offset"`
}

type TimeEntry struct {
	Id           int            `json:"id"`
	Hours        float64        `json:"hours"`
	Comments     string         `json:"comments"`
	SpentOn      string         `json:"spent_on"`
	Issue        *Name          `json:"issue"`
	Project      *Name          `json:"project"`
	Activity     *Name          `json:"activity"`
	User         *Name          `json:"user"`
	CustomFields []*CustomField `json:"custom_fields"`
	CreatedOn    string         `json:"created_on"`
	UpdatedOn    string         `json:"updated_on"`
}

type timeEntry struct {
	ProjectId    int            `json:"project_id,omitempty"`
	IssueId      int            `json:"issue_id,omitempty"`
	UserId       int            `json:"user_id,omitempty"`
	ActivityId   int            `json:"activity_id,omitempty"`
	Hours        float64        `json:"hours,omitempty"`
	Comments     string         `json:"comments,omitempty"`
	SpentOn      string         `json:"spent_on,omitempty"`
	CustomFields []*customField `json:"custom_fields,omitempty"`
}

func (r *TimeEntry) toSend() *timeEntry {
	newTimeEntry := new(timeEntry)
	if r.Project != nil {
		newTimeEntry.ProjectId = r.Project.Id
	}
	if r.Issue != nil {
		newTimeEntry.IssueId = r.Issue.Id
	}
	if r.User != nil {
		newTimeEntry.UserId = r.User.Id
	}
	if r.Activity != nil {
		newTimeEntry.ActivityId = r.Activity.Id
	}
	newTimeEntry.Hours = r.Hours
	newTimeEntry.Comments = r.Comments
	newTimeEntry.SpentOn = r.SpentOn
	if r.CustomFields != nil {
		newTimeEntry.CustomFields = make([]*customField, len(r.CustomFields))
		for i, cf := range r.CustomFields {
			newTimeEntry.CustomFields[i] = &customField{
				Id:    cf.Id,
				Value: cf.Value,
			}
		}
	}
	return newTimeEntry
}

//-------------------------------------------------------------------------
// list time entries
//-------------------------------------------------------------------------

type TimeEntriesListCall struct {
	s       *Service
	options map[string]interface{}
}

func (r *TimeEntriesService) List() *TimeEntriesListCall {
	return &TimeEntriesListCall{
		s:       r.s,
		options: make(map[string]interface{}),
	}
}

func (c *TimeEntriesListCall) Offset(offset int) *TimeEntriesListCall {
	c.options["offset"] = offset
	return c
}

func (c *TimeEntriesListCall) Limit(limit int) *TimeEntriesListCall {
	c.options["limit"] = limit
	return c
}

func (c *TimeEntriesListCall) Do() (*TimeEntryFeed, error) {
	params := make(url.Values)
	for _, opt := range []string{"offset", "limit"} {
		if v, ok := c.options[opt]; ok {
			params.Set(opt, fmt.Sprintf("%v", v))
		}
	}
	urlStr := resolveRelative(c.s.baseUrl, "time_entries.json")
	urlStr += "?" + params.Encode()
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := new(TimeEntryFeed)
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

//-------------------------------------------------------------------------
// get time entry
//-------------------------------------------------------------------------

type TimeEntriesGetCall struct {
	s           *Service
	timeEntryId int
}

func (r *TimeEntriesService) Get(timeEntryId int) *TimeEntriesGetCall {
	return &TimeEntriesGetCall{r.s, timeEntryId}
}

func (c *TimeEntriesGetCall) Do() (*TimeEntry, error) {
	urlStr := resolveRelative(c.s.baseUrl, "time_entries/{timeEntryId}.json")
	urlStr = strings.Replace(urlStr, "{timeEntryId}", strconv.Itoa(c.timeEntryId), 1)
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		TimeEntry *TimeEntry `json:"time_entry"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.TimeEntry, nil
}

//-------------------------------------------------------------------------
// insert time entry
//-------------------------------------------------------------------------

type TimeEntriesInsertCall struct {
	s         *Service
	timeEntry *TimeEntry
}

func (r *TimeEntriesService) Insert(timeEntry *TimeEntry) *TimeEntriesInsertCall {
	return &TimeEntriesInsertCall{r.s, timeEntry}
}

func (c *TimeEntriesInsertCall) Do() (*TimeEntry, error) {
	v := struct {
		TimeEntry *timeEntry `json:"time_entry,omitempty"`
	}{
		c.timeEntry.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return nil, err
	}
	urlStr := resolveRelative(c.s.baseUrl, "time_entries.json")
	data, err := c.s.doRequest("POST", urlStr, body)
	if err != nil {
		return nil, err
	}
	ret := struct {
		TimeEntry *TimeEntry `json:"time_entry"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.TimeEntry, nil
}

//-------------------------------------------------------------------------
// update time entry
//-------------------------------------------------------------------------

type TimeEntriesUpdateCall struct {
	s         *Service
	timeEntry *TimeEntry
}

func (r *TimeEntriesService) Update(timeEntry *TimeEntry) *TimeEntriesUpdateCall {
	return &TimeEntriesUpdateCall{r.s, timeEntry}
}

func (c *TimeEntriesUpdateCall) Do() error {
	v := struct {
		TimeEntry *timeEntry `json:"time_entry,omitempty"`
	}{
		c.timeEntry.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return err
	}
	urlStr := resolveRelative(c.s.baseUrl, "time_entries/{timeEntryId}.json")
	urlStr = strings.Replace(urlStr, "{timeEntryId}", strconv.Itoa(c.timeEntry.Id), 1)
	_, err = c.s.doRequest("PUT", urlStr, body)
	return err
}

//-------------------------------------------------------------------------
// delete time entry
//-------------------------------------------------------------------------

type TimeEntriesDeleteCall struct {
	s           *Service
	timeEntryId int
}

func (r *TimeEntriesService) Delete(timeEntryId int) *TimeEntriesDeleteCall {
	return &TimeEntriesDeleteCall{r.s, timeEntryId}
}

func (c *TimeEntriesDeleteCall) Do() error {
	urlStr := resolveRelative(c.s.baseUrl, "time_entries/{timeEntryId}.json")
	urlStr = strings.Replace(urlStr, "{timeEntryId}", strconv.Itoa(c.timeEntryId), 1)
	_, err := c.s.doRequest("DELETE", urlStr, nil)
	return err
}

//-------------------------------------------------------------------------
// news
//-------------------------------------------------------------------------

type NewsFeed struct {
	News       []*News `json:"news"`
	TotalCount int     `json:"total_count"`
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
}

type News struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Project     *Name  `json:"project"`
	Author      *Name  `json:"author"`
	CreatedOn   string `json:"created_on"`
}

//-------------------------------------------------------------------------
// list news
//-------------------------------------------------------------------------

type NewsListCall struct {
	s         *Service
	projectId int
	options   map[string]interface{}
}

func (r *NewsService) List() *NewsListCall {
	return &NewsListCall{
		s:       r.s,
		options: make(map[string]interface{}),
	}
}

func (c *NewsListCall) Offset(offset int) *NewsListCall {
	c.options["offset"] = offset
	return c
}

func (c *NewsListCall) Limit(limit int) *NewsListCall {
	c.options["limit"] = limit
	return c
}

func (c *NewsListCall) Project(projectId int) *NewsListCall {
	c.projectId = projectId
	return c
}

func (c *NewsListCall) Do() (*NewsFeed, error) {
	params := make(url.Values)
	for _, opt := range []string{"offset", "limit"} {
		if v, ok := c.options[opt]; ok {
			params.Set(opt, fmt.Sprintf("%v", v))
		}
	}
	var urlStr string
	if c.projectId == 0 {
		urlStr = resolveRelative(c.s.baseUrl, "news.json")
	} else {
		urlStr = resolveRelative(c.s.baseUrl, "projects/{projectId}/news.json")
		urlStr = strings.Replace(urlStr, "{projectId}", strconv.Itoa(c.projectId), 1)
	}
	urlStr += "?" + params.Encode()
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := new(NewsFeed)
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

//-------------------------------------------------------------------------
// relations
//-------------------------------------------------------------------------

type Relation struct {
	Id           int    `json:"id"`
	IssueId      int    `json:"issue_id"`
	IssueToId    int    `json:"issue_to_id"`
	RelationType string `json:"relation_type"`
	Delay        int    `json:"delay"`
}

type relation struct {
	IssueToId    int    `json:"issue_to_id,omitempty"`
	RelationType string `json:"relation_type,omitempty"`
	Delay        int    `json:"delay,omitempty"`
}

func (r *Relation) toSend() *relation {
	newRelation := new(relation)
	newRelation.IssueToId = r.IssueToId
	newRelation.RelationType = r.RelationType
	newRelation.Delay = r.Delay
	return newRelation
}

//-------------------------------------------------------------------------
// list relations
//-------------------------------------------------------------------------

type RelationsListCall struct {
	s       *Service
	issueId int
}

func (r *RelationsService) List(issueId int) *RelationsListCall {
	return &RelationsListCall{r.s, issueId}
}

func (c *RelationsListCall) Do() ([]*Relation, error) {
	urlStr := resolveRelative(c.s.baseUrl, "issues/{issueId}/relations.json")
	urlStr = strings.Replace(urlStr, "{issueId}", strconv.Itoa(c.issueId), 1)
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Relations []*Relation `json:"relations"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Relations, nil
}

//-------------------------------------------------------------------------
// get relation
//-------------------------------------------------------------------------

type RelationsGetCall struct {
	s          *Service
	relationId int
}

func (r *RelationsService) Get(relationId int) *RelationsGetCall {
	return &RelationsGetCall{r.s, relationId}
}

func (c *RelationsGetCall) Do() (*Relation, error) {
	urlStr := resolveRelative(c.s.baseUrl, "relations/{relationId}.json")
	urlStr = strings.Replace(urlStr, "{relationId}", strconv.Itoa(c.relationId), 1)
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Relation *Relation `json:"relation"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Relation, nil
}

//-------------------------------------------------------------------------
// insert relation
//-------------------------------------------------------------------------

type RelationsInsertCall struct {
	s        *Service
	relation *Relation
}

func (r *RelationsService) Insert(relation *Relation) *RelationsInsertCall {
	return &RelationsInsertCall{r.s, relation}
}

func (c *RelationsInsertCall) Do() (*Relation, error) {
	v := struct {
		Relation *relation `json:"relation,omitempty"`
	}{
		c.relation.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return nil, err
	}
	urlStr := resolveRelative(c.s.baseUrl, "issues/{issueId}/relations.json")
	urlStr = strings.Replace(urlStr, "{issueId}", strconv.Itoa(c.relation.IssueId), 1)
	data, err := c.s.doRequest("POST", urlStr, body)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Relation *Relation `json:"relation"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Relation, nil
}

//-------------------------------------------------------------------------
// delete relation
//-------------------------------------------------------------------------

type RelationsDeleteCall struct {
	s          *Service
	relationId int
}

func (r *RelationsService) Delete(relationId int) *RelationsDeleteCall {
	return &RelationsDeleteCall{r.s, relationId}
}

func (c *RelationsDeleteCall) Do() error {
	urlStr := resolveRelative(c.s.baseUrl, "relations/{relationId}.json")
	urlStr = strings.Replace(urlStr, "{relationId}", strconv.Itoa(c.relationId), 1)
	_, err := c.s.doRequest("DELETE", urlStr, nil)
	return err
}

//-------------------------------------------------------------------------
// versions
//-------------------------------------------------------------------------

// note: the custom fields are not being inserted or updated

type Version struct {
	Id           int            `json:"id"`
	Name         string         `json:"name"`
	Project      *Name          `json:"project"`
	Description  string         `json:"description"`
	Status       string         `json:"status"`
	Sharing      string         `json:"sharing"`
	DueDate      string         `json:"due_date"`
	CustomFields []*CustomField `json:"custom_fields"`
	CreatedOn    string         `json:"created_on"`
	UpdatedOn    string         `json:"updated_on"`
}

type version struct {
	Name         string         `json:"name,omitempty"`
	Description  string         `json:"description,omitempty"`
	Status       string         `json:"status,omitempty"`
	DueDate      string         `json:"due_date,omitempty"`
	Sharing      string         `json:"sharing,omitempty"`
	CustomFields []*customField `json:"custom_fields,omitempty"`
}

func (r *Version) toSend() *version {
	newVersion := new(version)
	newVersion.Name = r.Name
	newVersion.Description = r.Description
	newVersion.Status = r.Status
	newVersion.DueDate = r.DueDate
	newVersion.Sharing = r.Sharing
	if r.CustomFields != nil {
		newVersion.CustomFields = make([]*customField, len(r.CustomFields))
		for i, cf := range r.CustomFields {
			newVersion.CustomFields[i] = &customField{
				Id:    cf.Id,
				Value: cf.Value,
			}
		}
	}
	return newVersion
}

//-------------------------------------------------------------------------
// list versions
//-------------------------------------------------------------------------

type VersionsListCall struct {
	s         *Service
	projectId int
}

func (r *VersionsService) List(projectId int) *VersionsListCall {
	return &VersionsListCall{r.s, projectId}
}

func (c *VersionsListCall) Do() ([]*Version, error) {
	urlStr := resolveRelative(c.s.baseUrl, "projects/{projectId}/versions.json")
	urlStr = strings.Replace(urlStr, "{projectId}", strconv.Itoa(c.projectId), 1)
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Versions   []*Version `json:"versions"`
		TotalCount int        `json:"total_count"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Versions, nil
}

//-------------------------------------------------------------------------
// get version
//-------------------------------------------------------------------------

type VersionsGetCall struct {
	s         *Service
	versionId int
}

func (r *VersionsService) Get(versionId int) *VersionsGetCall {
	return &VersionsGetCall{r.s, versionId}
}

func (c *VersionsGetCall) Do() (*Version, error) {
	urlStr := resolveRelative(c.s.baseUrl, "versions/{versionId}.json")
	urlStr = strings.Replace(urlStr, "{versionId}", strconv.Itoa(c.versionId), 1)
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Version *Version `json:"version"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Version, nil
}

//-------------------------------------------------------------------------
// insert version
//-------------------------------------------------------------------------

type VersionsInsertCall struct {
	s       *Service
	version *Version
}

func (r *VersionsService) Insert(version *Version) *VersionsInsertCall {
	return &VersionsInsertCall{r.s, version}
}

func (c *VersionsInsertCall) Do() (*Version, error) {
	v := struct {
		Version *version `json:"version,omitempty"`
	}{
		c.version.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return nil, err
	}
	urlStr := resolveRelative(c.s.baseUrl, "projects/{versionId}/versions.json")
	urlStr = strings.Replace(urlStr, "{versionId}", strconv.Itoa(c.version.Project.Id), 1)
	data, err := c.s.doRequest("POST", urlStr, body)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Version *Version `json:"version"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Version, nil
}

//-------------------------------------------------------------------------
// update version
//-------------------------------------------------------------------------

type VersionsUpdateCall struct {
	s       *Service
	version *Version
}

func (r *VersionsService) Update(version *Version) *VersionsUpdateCall {
	return &VersionsUpdateCall{r.s, version}
}

func (c *VersionsUpdateCall) Do() error {
	v := struct {
		Version *version `json:"version,omitempty"`
	}{
		c.version.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return err
	}
	urlStr := resolveRelative(c.s.baseUrl, "versions/{versionId}.json")
	urlStr = strings.Replace(urlStr, "{versionId}", strconv.Itoa(c.version.Id), 1)
	_, err = c.s.doRequest("PUT", urlStr, body)
	return err
}

//-------------------------------------------------------------------------
// delete version
//-------------------------------------------------------------------------

type VersionsDeleteCall struct {
	s         *Service
	versionId int
}

func (r *VersionsService) Delete(versionId int) *VersionsDeleteCall {
	return &VersionsDeleteCall{r.s, versionId}
}

func (c *VersionsDeleteCall) Do() error {
	urlStr := resolveRelative(c.s.baseUrl, "versions/{versionId}.json")
	urlStr = strings.Replace(urlStr, "{versionId}", strconv.Itoa(c.versionId), 1)
	_, err := c.s.doRequest("DELETE", urlStr, nil)
	return err
}

//-------------------------------------------------------------------------
// wiki
//-------------------------------------------------------------------------

type WikiPage struct {
	Title       string        `json:"title"`
	Text        string        `json:"text"`
	Comments    string        `json:"comments"`
	Attachments []*Attachment `json:"attachments"`
	Version     int           `json:"version"`
	Author      *Name         `json:"author"`
	CreatedOn   string        `json:"created_on"`
	UpdatedOn   string        `json:"updated_on"`
}

type wikiPage struct {
	Title    string `json:"title,omitempty"`
	Text     string `json:"text,omitempty"`
	Comments string `json:"comments,omitempty"`
	Version  int    `json:"version,omitempty"`
}

func (r *WikiPage) toSend() *wikiPage {
	newWikiPage := new(wikiPage)
	newWikiPage.Title = r.Title
	newWikiPage.Text = r.Text
	newWikiPage.Comments = r.Comments
	newWikiPage.Version = r.Version
	return newWikiPage
}

//-------------------------------------------------------------------------
// list wiki pages
//-------------------------------------------------------------------------

type WikiListCall struct {
	s         *Service
	projectId int
}

func (r *WikiService) List(projectId int) *WikiListCall {
	return &WikiListCall{r.s, projectId}
}

func (c *WikiListCall) Do() ([]*WikiPage, error) {
	urlStr := resolveRelative(c.s.baseUrl, "projects/{projectId}/wiki/index.json")
	urlStr = strings.Replace(urlStr, "{projectId}", strconv.Itoa(c.projectId), 1)
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		WikiPages []struct {
			Title     string `json:"title"`
			Version   string `json:"version"`
			CreatedOn string `json:"created_on"`
			UpdatedOn string `json:"updated_on"`
		} `json:"wiki_pages"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	pages := make([]*WikiPage, len(ret.WikiPages))
	for i, p := range ret.WikiPages {
		page := new(WikiPage)
		page.Title = p.Title
		page.Version, _ = strconv.Atoi(p.Version)
		page.CreatedOn = p.CreatedOn
		page.UpdatedOn = p.UpdatedOn
		pages[i] = page
	}
	return pages, nil
}

//-------------------------------------------------------------------------
// get wiki page
//-------------------------------------------------------------------------

type WikiGetCall struct {
	s         *Service
	projectId int
	title     string
	version   int
	options   map[string]interface{}
}

func (r *WikiService) Get(projectId int, title string) *WikiGetCall {
	return &WikiGetCall{
		s:         r.s,
		projectId: projectId,
		title:     title,
		options:   make(map[string]interface{}),
	}
}

func (c *WikiGetCall) Version(version int) *WikiGetCall {
	c.version = version
	return c
}

func (c *WikiGetCall) Attachments(attachments bool) *WikiGetCall {
	c.options["attachments"] = attachments
	return c
}

func (c *WikiGetCall) Do() (*WikiPage, error) {
	params := make(url.Values)
	var include []string
	for _, inc := range []string{"attachments"} {
		v, ok := c.options[inc]
		if !ok {
			continue
		}
		if b, ok := v.(bool); ok && b {
			include = append(include, inc)
		}
	}
	if len(include) > 0 {
		params.Set("include", strings.Join(include, ","))
	}
	var urlStr string
	if c.version > 0 {
		urlStr = resolveRelative(c.s.baseUrl, "projects/{projectId}/wiki/{title}/{version}.json")
		urlStr = strings.Replace(urlStr, "{version}", strconv.Itoa(c.version), 1)
	} else {
		urlStr = resolveRelative(c.s.baseUrl, "projects/{projectId}/wiki/{title}.json")
	}
	urlStr = strings.Replace(urlStr, "{projectId}", strconv.Itoa(c.projectId), 1)
	urlStr = strings.Replace(urlStr, "{title}", c.title, 1)
	urlStr += "?" + params.Encode()
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		WikiPage *WikiPage `json:"wiki_page"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.WikiPage, nil
}

//-------------------------------------------------------------------------
// create or update a wiki page
//-------------------------------------------------------------------------

type WikiUpdateCall struct {
	s         *Service
	wikiPage  *WikiPage
	projectId int
}

func (r *WikiService) Update(wikiPage *WikiPage, projectId int) *WikiUpdateCall {
	return &WikiUpdateCall{r.s, wikiPage, projectId}
}

func (c *WikiUpdateCall) Do() error {
	v := struct {
		WikiPage *wikiPage `json:"wiki_page,omitempty"`
	}{
		c.wikiPage.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return err
	}
	urlStr := resolveRelative(c.s.baseUrl, "projects/{projectId}/wiki/{title}.json")
	urlStr = strings.Replace(urlStr, "{projectId}", strconv.Itoa(c.projectId), 1)
	urlStr = strings.Replace(urlStr, "{title}", c.wikiPage.Title, 1)
	_, err = c.s.doRequest("PUT", urlStr, body)
	return err
}

//-------------------------------------------------------------------------
// delete wiki page
//-------------------------------------------------------------------------

type WikiDeleteCall struct {
	s         *Service
	title     string
	projectId int
}

func (r *WikiService) Delete(title string, projectId int) *WikiDeleteCall {
	return &WikiDeleteCall{r.s, title, projectId}
}

func (c *WikiDeleteCall) Do() error {
	urlStr := resolveRelative(c.s.baseUrl, "projects/{projectId}/wiki/{title}.json")
	urlStr = strings.Replace(urlStr, "{projectId}", strconv.Itoa(c.projectId), 1)
	urlStr = strings.Replace(urlStr, "{title}", c.title, 1)
	_, err := c.s.doRequest("DELETE", urlStr, nil)
	return err
}

//-------------------------------------------------------------------------
// queries
//-------------------------------------------------------------------------

type QueryFeed struct {
	Queries    []*Query `json:"queries"`
	TotalCount int      `json:"total_count"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
}

type Query struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	IsPublic  bool   `json:"is_public"`
	ProjectId int    `json:"project_id"`
}

//-------------------------------------------------------------------------
// list queries
//-------------------------------------------------------------------------

type QueriesListCall struct {
	s       *Service
	options map[string]interface{}
}

func (r *QueriesService) List() *QueriesListCall {
	return &QueriesListCall{
		s:       r.s,
		options: make(map[string]interface{}),
	}
}

func (c *QueriesListCall) Offset(offset int) *QueriesListCall {
	c.options["offset"] = offset
	return c
}

func (c *QueriesListCall) Limit(limit int) *QueriesListCall {
	c.options["limit"] = limit
	return c
}

func (c *QueriesListCall) Do() (*QueryFeed, error) {
	params := make(url.Values)
	for _, opt := range []string{"offset", "limit"} {
		if v, ok := c.options[opt]; ok {
			params.Set(opt, fmt.Sprintf("%v", v))
		}
	}
	urlStr := resolveRelative(c.s.baseUrl, "queries.json")
	urlStr += "?" + params.Encode()
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := new(QueryFeed)
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

//-------------------------------------------------------------------------
// attachments
//-------------------------------------------------------------------------

type Attachment struct {
	Id          int    `json:"id"`
	Description string `json:"description"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Filesize    int    `json:"filesize"`
	ContentUrl  string `json:"content_url"`
	Author      *Name  `json:"author"`
	CreatedOn   string `json:"created_on"`
}

//-------------------------------------------------------------------------
// get attachment
//-------------------------------------------------------------------------

type AttachmentsGetCall struct {
	s            *Service
	attachmentId int
}

func (r *AttachmentsService) Get(attachmentId int) *AttachmentsGetCall {
	return &AttachmentsGetCall{r.s, attachmentId}
}

func (c *AttachmentsGetCall) Do() (*Attachment, error) {
	urlStr := resolveRelative(c.s.baseUrl, "attachments/{attachmentId}.json")
	urlStr = strings.Replace(urlStr, "{attachmentId}", strconv.Itoa(c.attachmentId), 1)
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Attachment *Attachment `json:"attachment"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Attachment, nil
}

//-------------------------------------------------------------------------
// issue statuses
//-------------------------------------------------------------------------

type IssueStatus struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	IsClosed  bool   `json:"is_closed"`
}

//-------------------------------------------------------------------------
// list issue statuses
//-------------------------------------------------------------------------

type IssueStatusesListCall struct {
	s *Service
}

func (r *IssueStatusesService) List() *IssueStatusesListCall {
	return &IssueStatusesListCall{r.s}
}

func (c *IssueStatusesListCall) Do() ([]*IssueStatus, error) {
	urlStr := resolveRelative(c.s.baseUrl, "issue_statuses.json")
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		IssueStatuses []*IssueStatus `json:"issue_statuses"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.IssueStatuses, nil
}

//-------------------------------------------------------------------------
// trackers
//-------------------------------------------------------------------------

type Tracker struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

//-------------------------------------------------------------------------
// list trackers
//-------------------------------------------------------------------------

type TrackersListCall struct {
	s *Service
}

func (r *TrackersService) List() *TrackersListCall {
	return &TrackersListCall{r.s}
}

func (c *TrackersListCall) Do() ([]*Tracker, error) {
	urlStr := resolveRelative(c.s.baseUrl, "trackers.json")
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Trackers []*Tracker `json:"trackers"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Trackers, nil
}

//-------------------------------------------------------------------------
// enumerations
//-------------------------------------------------------------------------

type Enumeration struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

//-------------------------------------------------------------------------
// enumerations: list document categories
//-------------------------------------------------------------------------

type DocumentCategoriesListCall struct {
	s *Service
}

func (r *DocumentCategoriesService) List() *DocumentCategoriesListCall {
	return &DocumentCategoriesListCall{r.s}
}

func (c *DocumentCategoriesListCall) Do() ([]*Enumeration, error) {
	urlStr := resolveRelative(c.s.baseUrl, "enumerations/document_categories.json")
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		DocumentCategories []*Enumeration `json:"document_categories"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.DocumentCategories, nil
}

//-------------------------------------------------------------------------
// enumerations: list issue priorities
//-------------------------------------------------------------------------

type IssuePrioritiesListCall struct {
	s *Service
}

func (r *IssuePrioritiesService) List() *IssuePrioritiesListCall {
	return &IssuePrioritiesListCall{r.s}
}

func (c *IssuePrioritiesListCall) Do() ([]*Enumeration, error) {
	urlStr := resolveRelative(c.s.baseUrl, "enumerations/issue_priorities.json")
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		IssuePriorities []*Enumeration `json:"issue_priorities"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.IssuePriorities, nil
}

//-------------------------------------------------------------------------
// enumerations: list time entry activities
//-------------------------------------------------------------------------

type TimeEntryActivitiesListCall struct {
	s *Service
}

func (r *TimeEntryActivitiesService) List() *TimeEntryActivitiesListCall {
	return &TimeEntryActivitiesListCall{r.s}
}

func (c *TimeEntryActivitiesListCall) Do() ([]*Enumeration, error) {
	urlStr := resolveRelative(c.s.baseUrl, "enumerations/time_entry_activities.json")
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		TimeEntryActivities []*Enumeration `json:"time_entry_activities"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.TimeEntryActivities, nil
}

//-------------------------------------------------------------------------
// issue categories
//-------------------------------------------------------------------------

type IssueCategory struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Project    *Name  `json:"project"`
	AssignedTo *Name  `json:"assigned_to"`
}

type issueCategory struct {
	Name         string `json:"name,omitempty"`
	AssignedToId int    `json:"assigned_to_id,omitempty"`
}

func (r *IssueCategory) toSend() *issueCategory {
	newIssueCategory := new(issueCategory)
	newIssueCategory.Name = r.Name
	if r.Project != nil {
		newIssueCategory.AssignedToId = r.Project.Id
	}
	return newIssueCategory
}

//-------------------------------------------------------------------------
// list issue categories
//-------------------------------------------------------------------------

type IssueCategoriesListCall struct {
	s         *Service
	projectId int
}

func (r *IssueCategoriesService) List(projectId int) *IssueCategoriesListCall {
	return &IssueCategoriesListCall{r.s, projectId}
}

func (c *IssueCategoriesListCall) Do() ([]*IssueCategory, error) {
	urlStr := resolveRelative(c.s.baseUrl, "projects/{projectId}/issue_categories.json")
	urlStr = strings.Replace(urlStr, "{projectId}", strconv.Itoa(c.projectId), 1)
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		IssueCategories []*IssueCategory `json:"issue_categories"`
		TotalCount      int              `json:"total_count"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.IssueCategories, nil
}

//-------------------------------------------------------------------------
// get issue category
//-------------------------------------------------------------------------

type IssueCategoriesGetCall struct {
	s               *Service
	issueCategoryId int
}

func (r *IssueCategoriesService) Get(issueCategoryId int) *IssueCategoriesGetCall {
	return &IssueCategoriesGetCall{r.s, issueCategoryId}
}

func (c *IssueCategoriesGetCall) Do() (*IssueCategory, error) {
	urlStr := resolveRelative(c.s.baseUrl, "issue_categories/{issueCategoryId}.json")
	urlStr = strings.Replace(urlStr, "{issueCategoryId}", strconv.Itoa(c.issueCategoryId), 1)
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		IssueCategory *IssueCategory `json:"issue_category"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.IssueCategory, nil
}

//-------------------------------------------------------------------------
// insert issue category
//-------------------------------------------------------------------------

type IssueCategoriesInsertCall struct {
	s             *Service
	issueCategory *IssueCategory
}

func (r *IssueCategoriesService) Insert(issueCategory *IssueCategory) *IssueCategoriesInsertCall {
	return &IssueCategoriesInsertCall{r.s, issueCategory}
}

func (c *IssueCategoriesInsertCall) Do() (*IssueCategory, error) {
	v := struct {
		IssueCategory *issueCategory `json:"issue_category,omitempty"`
	}{
		c.issueCategory.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return nil, err
	}
	urlStr := resolveRelative(c.s.baseUrl, "projects/{projectId}/issue_categories.json")
	urlStr = strings.Replace(urlStr, "{projectId}", strconv.Itoa(c.issueCategory.Project.Id), 1)
	data, err := c.s.doRequest("POST", urlStr, body)
	if err != nil {
		return nil, err
	}
	ret := struct {
		IssueCategory *IssueCategory `json:"issue_category"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.IssueCategory, nil
}

//-------------------------------------------------------------------------
// update issue category
//-------------------------------------------------------------------------

type IssueCategoriesUpdateCall struct {
	s             *Service
	issueCategory *IssueCategory
}

func (r *IssueCategoriesService) Update(issueCategory *IssueCategory) *IssueCategoriesUpdateCall {
	return &IssueCategoriesUpdateCall{r.s, issueCategory}
}

func (c *IssueCategoriesUpdateCall) Do() error {
	v := struct {
		IssueCategory *issueCategory `json:"issue_category,omitempty"`
	}{
		c.issueCategory.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return err
	}
	urlStr := resolveRelative(c.s.baseUrl, "issue_categories/{issueCategoryId}.json")
	urlStr = strings.Replace(urlStr, "{issueCategoryId}", strconv.Itoa(c.issueCategory.Id), 1)
	_, err = c.s.doRequest("PUT", urlStr, body)
	return err
}

//-------------------------------------------------------------------------
// delete issue category
//-------------------------------------------------------------------------

type IssueCategoriesDeleteCall struct {
	s               *Service
	issueCategoryId int
	reassignToId    int
}

func (r *IssueCategoriesService) Delete(issueCategoryId int) *IssueCategoriesDeleteCall {
	return &IssueCategoriesDeleteCall{
		s:               r.s,
		issueCategoryId: issueCategoryId,
	}
}

func (c *IssueCategoriesDeleteCall) ReassignTo(userId int) *IssueCategoriesDeleteCall {
	c.reassignToId = userId
	return c
}

func (c *IssueCategoriesDeleteCall) Do() error {
	params := make(url.Values)
	if c.reassignToId > 0 {
		params.Set("reassign_to_id", strconv.Itoa(c.reassignToId))
	}
	urlStr := resolveRelative(c.s.baseUrl, "issue_categories/{issueCategoryId}.json")
	urlStr = strings.Replace(urlStr, "{issueCategoryId}", strconv.Itoa(c.issueCategoryId), 1)
	urlStr += "?" + params.Encode()
	_, err := c.s.doRequest("DELETE", urlStr, nil)
	return err
}

//-------------------------------------------------------------------------
// roles
//-------------------------------------------------------------------------

type Role struct {
	Id          int      `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

//-------------------------------------------------------------------------
// list roles
//-------------------------------------------------------------------------

type RolesListCall struct {
	s *Service
}

func (r *RolesService) List() *RolesListCall {
	return &RolesListCall{r.s}
}

func (c *RolesListCall) Do() ([]*Role, error) {
	urlStr := resolveRelative(c.s.baseUrl, "roles.json")
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Roles []*Role `json:"roles"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Roles, nil
}

//-------------------------------------------------------------------------
// get role
//-------------------------------------------------------------------------

type RolesGetCall struct {
	s      *Service
	roleId int
}

func (r *RolesService) Get(roleId int) *RolesGetCall {
	return &RolesGetCall{r.s, roleId}
}

func (c *RolesGetCall) Do() (*Role, error) {
	urlStr := resolveRelative(c.s.baseUrl, "roles/{roleId}.json")
	urlStr = strings.Replace(urlStr, "{roleId}", strconv.Itoa(c.roleId), 1)
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Role *Role `json:"role"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Role, nil
}

//-------------------------------------------------------------------------
// groups
//-------------------------------------------------------------------------

type Group struct {
	Id           int                `json:"id"`
	Name         string             `json:"name"`
	Users        []*Name            `json:"users"`
	Memberships  []*GroupMembership `json:"memberships"`
	CustomFields []*CustomField     `json:"custom_fields"`
}

type GroupMembership struct {
	Id      int     `json:"id"`
	Project *Name   `json:"project"`
	Roles   []*Name `json:"roles"`
}

type group struct {
	Name         string         `json:"name,omitempty"`
	UserIds      []int          `json:"user_ids,omitempty"`
	CustomFields []*customField `json:"custom_fields,omitempty"`
}

func (r *Group) toSend() *group {
	newGroup := new(group)
	newGroup.Name = r.Name
	if r.Users != nil {
		newGroup.UserIds = make([]int, len(r.Users))
		for i, user := range r.Users {
			newGroup.UserIds[i] = user.Id
		}
	}
	if r.CustomFields != nil {
		newGroup.CustomFields = make([]*customField, len(r.CustomFields))
		for i, cf := range r.CustomFields {
			newGroup.CustomFields[i] = &customField{
				Id:    cf.Id,
				Value: cf.Value,
			}
		}
	}
	return newGroup
}

//-------------------------------------------------------------------------
// list groups
//-------------------------------------------------------------------------

type GroupsListCall struct {
	s *Service
}

func (r *GroupsService) List() *GroupsListCall {
	return &GroupsListCall{r.s}
}

func (c *GroupsListCall) Do() ([]*Group, error) {
	urlStr := resolveRelative(c.s.baseUrl, "groups.json")
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Groups []*Group `json:"groups"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Groups, nil
}

//-------------------------------------------------------------------------
// get group
//-------------------------------------------------------------------------

type GroupsGetCall struct {
	s       *Service
	groupId int
	options map[string]interface{}
}

func (r *GroupsService) Get(groupId int) *GroupsGetCall {
	return &GroupsGetCall{
		s:       r.s,
		groupId: groupId,
		options: make(map[string]interface{}),
	}
}

func (c *GroupsGetCall) Users(users bool) *GroupsGetCall {
	c.options["users"] = users
	return c
}

func (c *GroupsGetCall) Memberships(memberships bool) *GroupsGetCall {
	c.options["memberships"] = memberships
	return c
}

func (c *GroupsGetCall) Do() (*Group, error) {
	params := make(url.Values)
	var include []string
	for _, inc := range []string{"users", "memberships"} {
		v, ok := c.options[inc]
		if !ok {
			continue
		}
		if b, ok := v.(bool); ok && b {
			include = append(include, inc)
		}
	}
	if len(include) > 0 {
		params.Set("include", strings.Join(include, ","))
	}
	urlStr := resolveRelative(c.s.baseUrl, "groups/{groupId}.json")
	urlStr = strings.Replace(urlStr, "{groupId}", strconv.Itoa(c.groupId), 1)
	urlStr += "?" + params.Encode()
	data, err := c.s.doRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Group *Group `json:"group"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Group, nil
}

//-------------------------------------------------------------------------
// insert group
//-------------------------------------------------------------------------

type GroupsInsertCall struct {
	s     *Service
	group *Group
}

func (r *GroupsService) Insert(group *Group) *GroupsInsertCall {
	return &GroupsInsertCall{r.s, group}
}

func (c *GroupsInsertCall) Do() (*Group, error) {
	v := struct {
		Group *group `json:"group,omitempty"`
	}{
		c.group.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return nil, err
	}
	urlStr := resolveRelative(c.s.baseUrl, "groups.json")
	data, err := c.s.doRequest("POST", urlStr, body)
	if err != nil {
		return nil, err
	}
	ret := struct {
		Group *Group `json:"group"`
	}{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Group, nil
}

//-------------------------------------------------------------------------
// update group
//-------------------------------------------------------------------------

type GroupsUpdateCall struct {
	s     *Service
	group *Group
}

func (r *GroupsService) Update(group *Group) *GroupsUpdateCall {
	return &GroupsUpdateCall{r.s, group}
}

func (c *GroupsUpdateCall) Do() error {
	v := struct {
		Group *group `json:"group,omitempty"`
	}{
		c.group.toSend(),
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return err
	}
	urlStr := resolveRelative(c.s.baseUrl, "groups/{groupId}.json")
	urlStr = strings.Replace(urlStr, "{groupId}", strconv.Itoa(c.group.Id), 1)
	_, err = c.s.doRequest("PUT", urlStr, body)
	return err
}

//-------------------------------------------------------------------------
// delete group
//-------------------------------------------------------------------------

type GroupsDeleteCall struct {
	s       *Service
	groupId int
}

func (r *GroupsService) Delete(groupId int) *GroupsDeleteCall {
	return &GroupsDeleteCall{r.s, groupId}
}

func (c *GroupsDeleteCall) Do() error {
	urlStr := resolveRelative(c.s.baseUrl, "groups/{groupId}.json")
	urlStr = strings.Replace(urlStr, "{groupId}", strconv.Itoa(c.groupId), 1)
	_, err := c.s.doRequest("DELETE", urlStr, nil)
	return err
}

//-------------------------------------------------------------------------
// add user to group
//-------------------------------------------------------------------------

type GroupsAddUserCall struct {
	s       *Service
	groupId int
	userId  int
}

func (r *GroupsService) AddUser(groupId, userId int) *GroupsAddUserCall {
	return &GroupsAddUserCall{r.s, groupId, userId}
}

func (c *GroupsAddUserCall) Do() error {
	v := struct {
		UserId int `json:"user_id,omitempty"`
	}{
		c.userId,
	}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&v)
	if err != nil {
		return err
	}
	urlStr := resolveRelative(c.s.baseUrl, "groups/{groupId}/users.json")
	urlStr = strings.Replace(urlStr, "{groupId}", strconv.Itoa(c.groupId), 1)
	_, err = c.s.doRequest("POST", urlStr, body)
	return err
}

//-------------------------------------------------------------------------
// remove user from group
//-------------------------------------------------------------------------

type GroupsRemoveUserCall struct {
	s       *Service
	groupId int
	userId  int
}

func (r *GroupsService) RemoveUser(groupId, userId int) *GroupsRemoveUserCall {
	return &GroupsRemoveUserCall{r.s, groupId, userId}
}

func (c *GroupsRemoveUserCall) Do() error {
	urlStr := resolveRelative(c.s.baseUrl, "groups/{groupId}/users/{userId}.json")
	urlStr = strings.Replace(urlStr, "{groupId}", strconv.Itoa(c.groupId), 1)
	urlStr = strings.Replace(urlStr, "{userId}", strconv.Itoa(c.userId), 1)
	_, err := c.s.doRequest("DELETE", urlStr, nil)
	return err
}
