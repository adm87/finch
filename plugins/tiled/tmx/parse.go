package tmx

import "encoding/xml"

type Validator[T any] interface {
	Validate() error
}

func Parse[T Validator[T]](data []byte) (*T, error) {
	var m T
	if err := xml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, m.Validate()
}
