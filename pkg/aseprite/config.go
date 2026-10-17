package aseprite

import "github.com/adm87/finch/pkg/resources"

type LibraryID string

func (id LibraryID) String() string {
	return string(id)
}

func (id LibraryID) IsEmpty() bool {
	return id == ""
}

type Config struct {
	Libraries map[LibraryID]Library `yaml:"libraries"`
}

type Library struct {
	Manifest resources.Resource       `yaml:"manifest"`
	Content  resources.ResourceBundle `yaml:"content"`
}
