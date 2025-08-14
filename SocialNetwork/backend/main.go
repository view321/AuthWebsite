package main

import (
	"database/sql"
	"log"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
)

type NoteData struct{
	Text      string  `json:"text"`
	Longitude float64 `json:"longitude"`
	Lattitude float64 `json:"lattitude"`
}

var notes_db *sql.DB

func main() {
	dsn := "root:347347@tcp(127.0.0.1:3306)/notes_db"
	var err error
	notes_db, err = sql.Open("mysql", dsn)
	defer notes_db.Close()
	if err != nil{
		log.Fatal(err)
	}
	if err = notes_db.Ping(); err != nil{
		log.Fatal(err)
	}
	app := fiber.New()
	api := app.Group("/api")
	api.Post("/create_note", CreateGeoNote)
	api.Get("/get_by_id/:id", GetNoteByID)
	api.Delete("/delete_by_id/:id", DeleteNoteByID)
	api.Patch("/update_by_id/:id", UpdateNoteByID)
	log.Fatal(app.Listen(":8080"))
}

func CreateGeoNote(c *fiber.Ctx) error {
	var data NoteData
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error":"Invalid request"})
	}
	_, err := notes_db.Exec("INSERT INTO notes (note_text, longitude, lattitude) VALUES (?, ?, ?)", data.Text, data.Longitude, data.Lattitude)
	if err != nil{
		return c.Status(500).JSON(fiber.Map{"error" : "DB error"})
	}
	return c.JSON(fiber.Map{"message" : "Note added successfully"})
}

func GetNoteByID(c *fiber.Ctx) error{
	id_str := c.Params("id")
	id, err := strconv.Atoi(id_str)
	if err != nil{
		return c.Status(400).JSON(fiber.Map{"error":"Invalid request"})
	}
	var data NoteData
	row := notes_db.QueryRow("SELECT note_text, longitude, lattitude FROM notes WHERE id = ?", id)
	err = row.Scan(&data.Text, &data.Longitude, &data.Lattitude)
	if err != nil{
		return c.Status(400).JSON(fiber.Map{"error":"Invlaid database request"})
	}
	return c.JSON(data)
}

func DeleteNoteByID(c *fiber.Ctx) error{
	id_str := c.Params("id")
	id, err := strconv.Atoi(id_str)
	if err != nil{
		return c.Status(400).JSON(fiber.Map{"error":"Invalid request"})
	}
	_, err = notes_db.Exec("DELETE FROM notes WHERE id = ?", id)
	if err != nil{
		return c.Status(400).JSON(fiber.Map{"error":"Invlaid database request"})
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message":"Successfully deleted note"})
}

func UpdateNoteByID(c *fiber.Ctx) error{
	var data struct{
		Id int `json:"id"`
		Text      string  `json:"text"`
		Longitude float64 `json:"longitude"`
		Lattitude float64 `json:"lattitude"`
	}
	err := c.BodyParser(&data)
	if err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"wrong json"})
	}
	res, err := notes_db.Exec("UPDATE notes SET note_text=? longitude=? lattitude=? WHERE id=?", data.Text, data.Longitude, data.Lattitude, data.Id)
	if err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"database error"})
	}
	c.Status(fiber.StatusAccepted)
	c.JSON(res)
	return nil
}