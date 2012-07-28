package redmine

import (
	"encoding/json"
	"fmt"
	"time"
)

type Attachment struct {
	Id          int
	Author      *User
	Description string
	ContentUrl  string
	Filename    string
	Filesize    int
	ContentType string
	CreatedOn   time.Time
}

type attachmentDec struct {
	Id          *int     `json:"id"`
	Author      *userDec `json:"author"`
	Description *string  `json:"description"`
	ContentUrl  *string  `json:"content_url"`
	Filename    *string  `json:"filename"`
	Filesize    *int     `json:"filesize"`
	ContentType *string  `json:"content_type"`
	CreatedOn   *string  `json:"created_on"`
}

func (a *attachmentDec) decode() *Attachment {
	return &Attachment{
		Id:          ptrToInt(a.Id),
		Author:      a.Author.decode(),
		Description: ptrToString(a.Description),
		ContentUrl:  ptrToString(a.ContentUrl),
		Filename:    ptrToString(a.Filename),
		Filesize:    ptrToInt(a.Filesize),
		ContentType: ptrToString(a.ContentType),
		CreatedOn:   strToTime(TimeLayout, ptrToString(a.CreatedOn)),
	}
}

func unmarshalAttachment(data []byte) (*Attachment, error) {
	type response struct {
		Attachment *attachmentDec `json:"attachment"`
	}

	res := &response{}
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res.Attachment.decode(), nil
}

func (r *Redmine) GetAttachment(id int) (*Attachment, error) {
	path := fmt.Sprintf("attachments/%v.json", id)
	data, err := r.get(path, nil)
	if err != nil {
		return nil, err
	}

	return unmarshalAttachment(data)
}
