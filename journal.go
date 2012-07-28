package redmine

import "time"

type Journal struct {
	Id        int
	User      *User
	Details   []*JournalDetail
	Notes     string
	CreatedOn time.Time
}

type journalDec struct {
	Id        *int                `json:"id"`
	User      *userDec            `json:"user"`
	Details   []*journalDetailDec `json:"details"`
	Notes     *string             `json:"notes"`
	CreatedOn *string             `json:"created_on"`
}

func (j *journalDec) decode() *Journal {
	dec := &Journal{
		Id:        ptrToInt(j.Id),
		User:      j.User.decode(),
		Notes:     ptrToString(j.Notes),
		CreatedOn: strToTime(TimeLayout, ptrToString(j.CreatedOn)),
	}
	if j.Details != nil {
		dec.Details = make([]*JournalDetail, len(j.Details))
		for i := range j.Details {
			dec.Details[i] = j.Details[i].decode()
		}
	}

	return dec
}
