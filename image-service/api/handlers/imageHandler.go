package handlers

import (
	"errors"
	"fmt"
	"image-service/pkg/core"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())

var extFileTypes map[string]string = map[string]string{
	"jpeg": "image/jpeg",
	"jpg":  "image/jpg",
	"png":  "image/png",
	"webp": "image/webp",
}

func GetImageFile(service core.ImageService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fileName := c.Params("name")
		extName := strings.ToLower(strings.Replace(filepath.Ext(fileName), ".", "", -1))
		fmt.Println(extName)

		if _, prs := extFileTypes[extName]; !prs {
			c.Status(http.StatusBadRequest)
			return c.JSON(GetErrorResponse(errors.New("Invalid file extension")))
		}

		imageFile, err := service.GetImageFile(fileName)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(GetErrorResponse(err))
		}

		c.Set("Content-Type", extFileTypes[extName])
		return c.Send(imageFile)
	}
}

func GetImage(service core.ImageService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		image, err := service.GetImage(c.Params("id"))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(GetErrorResponse(err))
		}

		return c.JSON(image)
	}
}

func AddImage(service core.ImageService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var requestBody ImageCreateRequestDto
		err := c.BodyParser(&requestBody)

		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(GetErrorResponse(err))
		}

		validationErr := validateStruct(requestBody)

		if validationErr != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(
				GetErrorResponse(validationErr),
			)
		}

		fileHeader, err := c.FormFile("image")

		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(GetErrorResponse(err))
		}

		file, err := formFileToBytes(fileHeader)

		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(GetErrorResponse(err))
		}

		if requestBody.AvailableFormats == nil {
			requestBody.AvailableFormats = make([]string, 0)
		}

		imageCreateDto := core.ImageCreateDto{
			Name: requestBody.Name,
			// Url:              requestBody.Url,
			AvailableFormats: requestBody.AvailableFormats,
			File:             *file,
			OriginalName:     &fileHeader.Filename,
		}

		image, err := service.CreateImage(imageCreateDto, true)

		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(GetErrorResponse(err))
		}

		return c.JSON(image)
	}
}

func UpdateImage(service core.ImageService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var requestBody ImageUpdateRequestDto
		err := c.BodyParser(&requestBody)

		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(GetErrorResponse(err))
		}

		validationErr := validateStruct(requestBody)

		if validationErr != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(
				GetErrorResponse(validationErr),
			)
		}

		fileHeader, err := c.FormFile("image")
		var file *[]byte
		var filename *string

		if fileHeader != nil {
			filename = &fileHeader.Filename

			file, err = formFileToBytes(fileHeader)

			if err != nil {
				c.Status(http.StatusBadRequest)
				return c.JSON(GetErrorResponse(err))
			}
		}

		imageUpdateDto := core.ImageUpdateDto{
			Name: requestBody.Name,
			AvailableFormats: &requestBody.AvailableFormats,
			File:             file,
			OriginalName:     filename,
		}

		image, err := service.UpdateImage(c.Params("id"), imageUpdateDto, true)

		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(GetErrorResponse(err))
		}

		return c.JSON(image)
	}
}

func DeleteImage(service core.ImageService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		image, err := service.DeleteImage(c.Params("id"))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(GetErrorResponse(err))
		}

		return c.JSON(image)
	}
}

func validateStruct(data interface{}) error {
	// returns nil or ValidationErrors ( []FieldError )
	err := validate.Struct(data)
	if err != nil {

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
			return err
		}

		var errMessage = ""

		for _, err := range err.(validator.ValidationErrors) {
			errMessage = fmt.Sprintf(
				"Error in %s field: error tag - %s, error value - %s.",
				err.Field(),
				err.Tag(),
				err.Value(),
			)

			// fmt.Println(err.Namespace())
			// fmt.Println(err.Field())
			// fmt.Println(err.StructNamespace())
			// fmt.Println(err.StructField())
			// fmt.Println(err.Tag())
			// fmt.Println(err.ActualTag())
			// fmt.Println(err.Kind())
			// fmt.Println(err.Type())
			// fmt.Println(err.Value())
			// fmt.Println(err.Param())
			// fmt.Println()
		}

		return errors.New(errMessage)
	}

	return nil
}
