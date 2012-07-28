package redmine

import (
	"encoding/json"
	"fmt"
)

type Upload struct {
	Token       string `json:"token"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

type uploadDec struct {
	Token *string `json:"token,omitempty"`
}

func (u *uploadDec) decode() *Upload {
	return &Upload{
		Token: ptrToString(u.Token),
	}
}

func unmarshalUpload(data []byte) (*Upload, error) {
	type response struct {
		Upload *uploadDec `json:"upload"`
	}

	res := &response{}
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res.Upload.decode(), nil
}

func (r *Redmine) UploadFile(file []byte) (*Upload, error) {
	rawurl, err := buildUrl(r.url, "uploads.json", nil)
	if err != nil {
		return nil, err
	}

	data, status, err := r.sendRequest("POST", rawurl, "application/octet-stream", file)
	if err != nil {
		return nil, err
	}

	if status != "201 Created" {
		return nil, fmt.Errorf("%s %s", status, string(data))
	}

	return unmarshalUpload(data)
}
