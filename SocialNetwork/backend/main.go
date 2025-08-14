package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	api := app.Group("/api")
	api.Post("/create_note", CreateGeoNote)
	log.Fatal(app.Listen(":8080"))
}

func CreateGeoNote(c *fiber.Ctx) error {
	var data struct {
		Longitude float64 `json:"longitude"`
		Lattitude float64 `json:"lattitude"`
		Text      string  `json:"text"`
	}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON
	}
	return nil
}
