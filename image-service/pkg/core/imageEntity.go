package core

import "time"

type ImageEntity struct {
	Id               string    `json:"id"`
	Name             string    `json:"name"`
	Url              string    `json:"url"`
	CreatedDate      time.Time `json:"createdDate"`
	UpdatedDate      time.Time `json:"updatedDate"`
	AvailableFormats []string  `json:"availableFormats"`
}
