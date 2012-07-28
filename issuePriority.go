package redmine

type IssuePriority struct {
	Id           int
	Name         string
	CustomFields []*CustomField
}

type issuePriorityDec struct {
	Id           *int              `json:"id"`
	Name         *string           `json:"name"`
	CustomFields []*customFieldDec `json:"custom_fields"`
}

func (i *issuePriorityDec) decode() *IssuePriority {
	dec := &IssuePriority{
		Id:   ptrToInt(i.Id),
		Name: ptrToString(i.Name),
	}
	if i.CustomFields != nil {
		dec.CustomFields = make([]*CustomField, len(i.CustomFields))
		for j := range i.CustomFields {
			dec.CustomFields[j] = i.CustomFields[j].decode()
		}
	}

	return dec
}
