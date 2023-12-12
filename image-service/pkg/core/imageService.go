package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ImageService interface {
	GetImage(id string) (*ImageEntity, error)
	DeleteImage(id string) (int, error)
	CreateImage(image ImageCreateDto, isAsync bool) (*ImageEntity, error)
	UpdateImage(id string, image ImageUpdateDto, isAsync bool) (*ImageEntity, error)
	GetImageFile(name string) ([]byte, error)
}

type imageService struct {
	repository  ImageRepository
	dataStorage DataStorage
	appHost     string
}

func NewImageService(r ImageRepository, dataStorage DataStorage, appHost string) ImageService {
	return &imageService{
		repository:  r,
		dataStorage: dataStorage,
		appHost:     appHost,
	}
}

func (s *imageService) GetImage(id string) (*ImageEntity, error) {
	return s.repository.GetImageById(id)
}

func (s *imageService) GetImageFile(name string) ([]byte, error) {
	return s.dataStorage.GetFile(name)
}

func (s *imageService) DeleteImage(id string) (int, error) {
	image, err := s.repository.GetImageById(id)

	if err != nil {
		return 0, errors.New("Image not found")
	}

	err = s.dataStorage.DeleteImage(image.Id, image.AvailableFormats)

	if err != nil {
		return 0, err
	}

	return s.repository.DeleteImageById(image.Id)
}

func (s *imageService) CreateImage(imageDto ImageCreateDto, isAsync bool) (*ImageEntity, error) {
	uuid := uuid.New().String()
	imageDto.Id = &uuid

	if imageDto.Name == nil {
		imageDto.Name = imageDto.Id
	}

	var err error

	if isAsync {
		err = s.dataStorage.SaveImageAsync(imageDto.File, *imageDto.OriginalName, *imageDto.Id, imageDto.AvailableFormats)
	} else {
		err = s.dataStorage.SaveImage(imageDto.File, *imageDto.Id, imageDto.AvailableFormats)
	}

	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/get-file/%s.%s", s.appHost, *imageDto.Id, imageDto.AvailableFormats[0])
	imageDto.Url = &url

	return s.repository.CreateImage(imageDto)
}

func (s *imageService) UpdateImage(id string, imageDto ImageUpdateDto, isAsync bool) (*ImageEntity, error) {
	image, err := s.repository.GetImageById(id)

	if err != nil {
		return nil, errors.New("Image not found")
	}

	if imageDto.Name != nil {
		image.Name = *imageDto.Name
	}

	if imageDto.File != nil {
		err = s.dataStorage.DeleteImage(image.Id, image.AvailableFormats)

		if err != nil {
			return nil, err
		}

		var availableFormats []string

		if imageDto.AvailableFormats == nil {
			availableFormats = image.AvailableFormats
		} else {
			availableFormats = *imageDto.AvailableFormats
		}

		image.AvailableFormats = availableFormats

		err = s.dataStorage.SaveImageAsync(*imageDto.File, *imageDto.OriginalName, image.Id, availableFormats)

		if err != nil {
			return nil, err
		}

		url := fmt.Sprintf("%s/api/get-file/%s.%s", s.appHost, image.Id, image.AvailableFormats[0])
		imageDto.Url = &url
	}

	// image.UpdatedDate = time.Now().Format(time.RFC3339)
	image.UpdatedDate = time.Now()

	return s.repository.UpdateImage(*image)
}
