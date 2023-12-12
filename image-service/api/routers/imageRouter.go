package routers

import (
	"image-service/api/handlers"
	"image-service/pkg/core"

	"github.com/gofiber/fiber/v2"
)

func ImageRouter(app fiber.Router, service core.ImageService) {
	app.Get("/image/:id", handlers.GetImage(service))
	app.Get("/get-file/:name", handlers.GetImageFile(service))
	app.Post("/image", handlers.AddImage(service))
	app.Patch("/image/:id", handlers.UpdateImage(service))
	app.Delete("/image/:id", handlers.DeleteImage(service))
}
