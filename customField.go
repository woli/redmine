package redmine

import (
	"fmt"
)

type CustomField struct {
	Id       int
	Name     string
	Value    string
	Multiple bool
	Values   []string
}

type customFieldDec struct {
	Id       *int        `json:"id"`
	Name     *string     `json:"name"`
	Multiple *bool       `json:"multiple"`
	Value    interface{} `json:"value"`
}

func (c *customFieldDec) decode() *CustomField {
	dec := &CustomField{
		Id:       ptrToInt(c.Id),
		Name:     ptrToString(c.Name),
		Multiple: ptrToBool(c.Multiple),
	}

	if dec.Multiple {
		values, ok := c.Value.([]interface{})
		if !ok {
			return dec
		}

		dec.Values = make([]string, 0)
		for i := range values {
			value, ok := values[i].(string)
			if ok {
				dec.Values = append(dec.Values, value)
			}
		}

	} else {
		dec.Value, _ = c.Value.(string)
	}

	return dec
}

func customFieldsToMap(customFields []*CustomField) map[string]interface{} {
	m := make(map[string]interface{})
	for i := range customFields {
		if customFields[i].Multiple {
			m[fmt.Sprint(customFields[i].Id)] = customFields[i].Values
		} else {
			m[fmt.Sprint(customFields[i].Id)] = customFields[i].Value
		}
	}

	return m
}
