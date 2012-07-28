package redmine

type TimeActivity struct {
	Id           int
	Name         string
	CustomFields []*CustomField
}

type timeActivityDec struct {
	Id           *int              `json:"id"`
	Name         *string           `json:"name"`
	CustomFields []*customFieldDec `json:"custom_fields"`
}

func (t *timeActivityDec) decode() *TimeActivity {
	dec := &TimeActivity{
		Id:   ptrToInt(t.Id),
		Name: ptrToString(t.Name),
	}
	if t.CustomFields != nil {
		dec.CustomFields = make([]*CustomField, len(t.CustomFields))
		for i := range t.CustomFields {
			dec.CustomFields[i] = t.CustomFields[i].decode()
		}
	}

	return dec
}
