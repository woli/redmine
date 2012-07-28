package redmine

import (
	"encoding/json"
	"net/url"
)

type Tracker struct {
	Id   int
	Name string
}

type trackerDec struct {
	Id   *int    `json:"id"`
	Name *string `json:"name"`
}

func (t *trackerDec) decode() *Tracker {
	return &Tracker{
		Id:   ptrToInt(t.Id),
		Name: ptrToString(t.Name),
	}
}

func (r *Redmine) GetTrackers(v *url.Values) ([]*Tracker, *Pagination, error) {
	data, err := r.get("trackers.json", v)
	if err != nil {
		return nil, nil, err
	}

	type response struct {
		Trackers   []*trackerDec `json:"trackers"`
		TotalCount int           `json:"total_count"`
		Limit      int           `json:"limit"`
		Offset     int           `json:"offset"`
	}

	res := &response{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, nil, err
	}

	trackers := make([]*Tracker, len(res.Trackers))
	for i := range res.Trackers {
		trackers[i] = res.Trackers[i].decode()
	}

	pagination := &Pagination{res.TotalCount, res.Limit, res.Offset}
	return trackers, pagination, nil
}
