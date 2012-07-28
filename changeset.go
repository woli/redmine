package redmine

import "time"

type Changeset struct {
	Revision   string
	User       *User
	Comments   string
	CommitedOn time.Time
}

type changesetDec struct {
	Revision   *string  `json:"revision"`
	User       *userDec `json:"user"`
	Comments   *string  `json:"comments"`
	CommitedOn *string  `json:"commited_on"`
}

func (c *changesetDec) decode() *Changeset {
	return &Changeset{
		Revision:   ptrToString(c.Revision),
		User:       c.User.decode(),
		Comments:   ptrToString(c.Comments),
		CommitedOn: strToTime(TimeLayout, ptrToString(c.CommitedOn)),
	}
}
