package handlers

import (
	"bytes"
	"io"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
)

func GetErrorResponse(err error) *fiber.Map {
	return &fiber.Map{
		"status": false,
		"data":   "",
		"error":  err.Error(),
	}
}

func formFileToBytes(fileHeader *multipart.FileHeader) (*[]byte, error) {
	file, err := fileHeader.Open()

	if err != nil {
		return nil, err
	}

	defer file.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return nil, err
	}

	bytes := buf.Bytes()
	return &bytes, nil
}
