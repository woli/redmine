package redmine

type JournalDetail struct {
	Name     string
	Property string
	NewValue string
}

type journalDetailDec struct {
	Name     *string `json:"name"`
	Property *string `json:"property"`
	NewValue *string `json:"new_value"`
}

func (j *journalDetailDec) decode() *JournalDetail {
	return &JournalDetail{
		Name:     ptrToString(j.Name),
		Property: ptrToString(j.Property),
		NewValue: ptrToString(j.NewValue),
	}
}
