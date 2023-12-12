package handlers

type ImageCreateRequestDto struct {
	Name             *string  `json:"name,omitempty" validate:"omitempty"`
	AvailableFormats []string `json:"availableFormats,omitempty" validate:"unique,dive,oneof=png jpg jpeg webp avif"`
}

type ImageUpdateRequestDto struct {
	Name             *string  `json:"name,omitempty" validate:"omitempty"`
	AvailableFormats []string `json:"availableFormats,omitempty" validate:"omitempty,unique,dive,oneof=png jpg jpeg webp avif"`
}
