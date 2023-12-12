package imageProcessor

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"runtime"

	"github.com/google/uuid"
)

var supportedFileTypes map[string]string = map[string]string{
	"jpeg": "jpeg",
	"jpg":  "jpg",
	"png":  "png",
	"webp": "webp",
	"avif": "avif",
}

type ImageData struct {
	Name string
	File []byte
}

type convert func(fullOriginalFileName string, fullConvertedFileName string) error

type ImageProcessor struct{}

func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{}
}

func (ip *ImageProcessor) ConvertImage(file []byte, originalName string, name string, format string) (*ImageData, error) {
	imgBuf := bytes.NewBuffer(file)
	imgDecoded, _, err := image.Decode(imgBuf)

	if err != nil {
		return nil, err
	}

	var encodedBuf bytes.Buffer
	w := bufio.NewWriter(&encodedBuf)
	var convertedFile []byte

	switch format {
	case "jpg", "jpeg":
		err = jpeg.Encode(w, imgDecoded, nil)
	case "png":
		err = png.Encode(w, imgDecoded)
	case "webp":
		convertedFile, err = convertInShell(
			file,
			originalName,
			"webp",
			func(fullOriginalFileName string, fullConvertedFileName string) error {
				cmd := exec.Command("cwebp", fullOriginalFileName, "-o", fullConvertedFileName)
				_, err := cmd.Output()
				return err
			},
		)
	case "avif":
		convertedFile, err = convertInShell(
			file,
			originalName,
			"avif",
			func(fullOriginalFileName string, fullConvertedFileName string) error {
				cmd := exec.Command("convert", fullOriginalFileName, fullConvertedFileName)
				_, err := cmd.Output()
				return err
			},
		)
	default:
		err = errors.New(fmt.Sprintf("Формат %s не поддерживается", format))
	}

	if err != nil {
		return nil, err
	}

	if len(convertedFile) == 0 {
		convertedFile = encodedBuf.Bytes()
	}

	return &ImageData{
			Name: name + "." + format,
			File: convertedFile,
		},
		nil
}

func convertInShell(file []byte, originalName string, convertFormat string, convertFn convert) ([]byte, error) {
	if runtime.GOOS != "linux" {
		return []byte{}, errors.New("Runtime OS isn't Linux - webp and avif conversion is not supported")
	}

	fullOriginalFileName := fmt.Sprintf("/tmp/%w", originalName)

	if err := os.WriteFile(fullOriginalFileName, file, 0644); err != nil {
		return []byte{}, err
	}

	fullConvertedFileName := fmt.Sprintf("/tmp/%w.%w", uuid.New().String(), convertFormat)

	err := convertFn(fullOriginalFileName, fullConvertedFileName)

	if err != nil {
		return []byte{}, err
	}

	convertedFile, err := os.ReadFile(fullConvertedFileName)

	if err != nil {
		return []byte{}, err
	}

	os.Remove(fullOriginalFileName)
	os.Remove(fullConvertedFileName)

	return convertedFile, nil
}
