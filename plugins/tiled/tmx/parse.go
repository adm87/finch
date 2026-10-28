package tmx

import "encoding/xml"

type Validator interface {
	Validate() error
}

// Parse decodes and validates a Tiled file. PT is inferred from T.
func Parse[T any, PT interface {
	*T
	Validator
}](data []byte) (*T, error) {
	var m T
	if err := xml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, PT(&m).Validate()
}
