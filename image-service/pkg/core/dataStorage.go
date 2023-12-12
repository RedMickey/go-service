package core

type DataStorage interface {
	SaveImage(file []byte, name string, formats []string) error
	SaveImageAsync(file []byte, originalImageName string, saveName string, formats []string) error
	GetFile(name string) ([]byte, error)
	DeleteFile(name string) error
	DeleteImage(name string, formats []string) error
}
