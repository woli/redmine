package redmine

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var TimeLayout = "2006-01-02T15:04:05Z"
var DateLayout = "2006-01-02"

// Sets the certificate authority to be used in all API requests.
func SetCertAuth(pem []byte) {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(pem)
	http.DefaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: pool,
		},
	}
}

type Redmine struct {
	url      string
	login    string
	password string
}

type Pagination struct {
	TotalCount int
	Limit      int
	Offset     int
}

func New(url string) *Redmine {
	return &Redmine{url: url}
}

func (r *Redmine) SetAPIKey(key string) {
	r.SetBasicAuth(key, "password")
}

func (r *Redmine) SetBasicAuth(login, password string) {
	r.login = login
	r.password = password
}

func (r *Redmine) sendRequest(method, url string, contentType string, body []byte) (data []byte, status string, err error) {
	var reader io.Reader = nil
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	req.SetBasicAuth(r.login, r.password)
	req.ContentLength = int64(len(body))
	req.Header.Add("Content-Type", contentType)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}

	defer res.Body.Close()

	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, "", err
	}

	return data, strings.Trim(res.Status, " "), nil
}

func (r *Redmine) get(path string, v *url.Values) ([]byte, error) {
	rawurl, err := buildUrl(r.url, path, v)
	if err != nil {
		return nil, err
	}

	data, status, err := r.sendRequest("GET", rawurl, "application/json", nil)
	if err != nil {
		return nil, err
	}

	if status != "200 OK" {
		return nil, fmt.Errorf("%s %s", status, string(data))
	}

	return data, nil
}

func (r *Redmine) post(path string, body []byte) ([]byte, error) {
	rawurl, err := buildUrl(r.url, path, nil)
	if err != nil {
		return nil, err
	}

	data, status, err := r.sendRequest("POST", rawurl, "application/json", body)
	if err != nil {
		return nil, err
	}

	if status != "201 Created" {
		return nil, fmt.Errorf("%s %s", status, string(data))
	}

	return data, nil
}

func (r *Redmine) put(path string, body []byte) error {
	rawurl, err := buildUrl(r.url, path, nil)
	if err != nil {
		return err
	}

	data, status, err := r.sendRequest("PUT", rawurl, "application/json", body)
	if err != nil {
		return err
	}

	if status != "200 OK" {
		return fmt.Errorf("%s %s", status, string(data))
	}

	return nil
}

func (r *Redmine) delete(path string) error {
	rawurl, err := buildUrl(r.url, path, nil)
	if err != nil {
		return err
	}

	data, status, err := r.sendRequest("DELETE", rawurl, "application/json", nil)
	if err != nil {
		return err
	}

	if status != "200 OK" {
		return fmt.Errorf("%s %s", status, string(data))
	}

	return nil
}

func buildUrl(host, path string, v *url.Values) (string, error) {
	rawurl := fmt.Sprintf("%s/%s", host, path)
	uri, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}

	if v != nil {
		return fmt.Sprintf("%s?%s", uri.String(), v.Encode()), nil
	}

	return uri.String(), nil
}

func ptrToString(v *string) string {
	if v != nil {
		return *v
	}

	return ""
}

func ptrToInt(v *int) int {
	if v != nil {
		return *v
	}

	return 0
}

func ptrToFloat64(v *float64) float64 {
	if v != nil {
		return *v
	}

	return 0
}

func ptrToBool(v *bool) bool {
	if v != nil {
		return *v
	}

	return false
}

func strToTime(layout string, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		return time.Time{}
	}

	return t
}

func timeToStr(layout string, value time.Time) string {
	defaultTime := time.Time{}
	if value == defaultTime {
		return ""
	}

	return value.Format(layout)
}
