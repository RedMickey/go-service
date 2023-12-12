package core

type ImageRepository interface {
	GetImageById(id string) (*ImageEntity, error)
	DeleteImageById(id string) (int, error)
	CreateImage(image ImageCreateDto) (*ImageEntity, error)
	UpdateImage(image ImageEntity) (*ImageEntity, error)
}
