package redmine

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type News struct {
	Id          int
	Project     *Project
	Author      *User
	Title       string
	Summary     string
	Description string
	CreatedOn   time.Time
}

type newsDec struct {
	Id          *int        `json:"id"`
	Project     *projectDec `json:"project"`
	Author      *userDec    `json:"author"`
	Title       *string     `json:"title"`
	Summary     *string     `json:"summary"`
	Description *string     `json:"description"`
	CreatedOn   *string     `json:"created_on"`
}

func (n *newsDec) decode() *News {
	dec := &News{
		Id:          ptrToInt(n.Id),
		Title:       ptrToString(n.Title),
		Summary:     ptrToString(n.Summary),
		Description: ptrToString(n.Description),
		CreatedOn:   strToTime(TimeLayout, ptrToString(n.CreatedOn)),
	}
	if n.Project != nil {
		dec.Project = n.Project.decode()
	}
	if n.Author != nil {
		dec.Author = n.Author.decode()
	}

	return dec
}

func (r *Redmine) GetNews(v *url.Values) ([]*News, *Pagination, error) {
	data, err := r.get("news.json", v)
	if err != nil {
		return nil, nil, err
	}

	type response struct {
		News       []*newsDec `json:"news"`
		TotalCount int        `json:"total_count"`
		Limit      int        `json:"limit"`
		Offset     int        `json:"offset"`
	}

	res := &response{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, err
	}

	news := make([]*News, len(res.News))
	for i := range res.News {
		news[i] = res.News[i].decode()
	}

	pagination := &Pagination{res.TotalCount, res.Limit, res.Offset}
	return news, pagination, nil
}

func (r *Redmine) GetProjectNews(projectId int, v *url.Values) ([]*News, *Pagination, error) {
	path := fmt.Sprintf("projects/%v/news.json", projectId)
	data, err := r.get(path, v)
	if err != nil {
		return nil, nil, err
	}

	type response struct {
		News       []*newsDec `json:"news"`
		TotalCount int        `json:"total_count"`
		Limit      int        `json:"limit"`
		Offset     int        `json:"offset"`
	}

	res := &response{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, err
	}

	news := make([]*News, len(res.News))
	for i := range res.News {
		news[i] = res.News[i].decode()
	}

	pagination := &Pagination{res.TotalCount, res.Limit, res.Offset}
	return news, pagination, nil
}
