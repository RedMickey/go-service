package core

type ImageCreateDto struct {
	Id               *string
	Name             *string
	Url              *string
	AvailableFormats []string
	File             []byte
	OriginalName     *string
}

type ImageUpdateDto struct {
	Name             *string
	Url              *string
	AvailableFormats *[]string
	File             *[]byte
	OriginalName     *string
}
