package main

import (
	"log"
	"database/sql"

	"github.com/gofiber/fiber/v2"
	_ "github.com/go-sql-driver/mysql"
)

var notes_db *sql.DB

func main() {
	dsn := "root:347347@tcp(127.0.0.1:3306)/notes_db"
	db, err := sql.Open("mysql", dsn)
	if err != nil{
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil{
		log.Fatal(err)
	}
	app := fiber.New()
	api := app.Group("/api")
	api.Post("/create_note", CreateGeoNote)
	api.Get("/get_by_id/:id", GetNoteByID)
	log.Fatal(app.Listen(":8080"))
}

func CreateGeoNote(c *fiber.Ctx) error {
	var data struct {
		Text      string  `json:"text"`
		Longitude float64 `json:"longitude"`
		Lattitude float64 `json:"lattitude"`
	}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error":"Invalid request"})
	}
	_, err := notes_db.Exec("INSERT INTO notes (text, longitude, lattitude) VALUES (?, ?, ?)", data.Text, data.Longitude, data.Lattitude)
	if err != nil{
		return c.Status(500).JSON(fiber.Map{"error" : "DB error"})
	}
	return c.JSON(fiber.Map{"message" : "Note added successfully"})
}

func GetNoteByID(c *fiber.Ctx) error{
	id := c.Params("id")
	
}