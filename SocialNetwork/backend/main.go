package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type NoteData struct{
	Text      string  `json:"text"`
	Longitude float64 `json:"longitude"`
	Lattitude float64 `json:"lattitude"`
	UserID string `json:"user_id"`
}

var notes_db *sql.DB
var jwtSecret = []byte("my_key_123")
var blacklist = make(map[string]bool)

func main() {
	dsn := "root:347347@tcp(127.0.0.1:3306)/notes_db"
	var err error
	notes_db, err = sql.Open("mysql", dsn)	
	if err != nil{
		log.Fatal(err)
	}
	defer notes_db.Close()
	if err = notes_db.Ping(); err != nil{
		log.Fatal(err)
	}
	app := fiber.New()
	app.Use(logger.New())
	api := app.Group("/api")

	login_protected := api.Group("/login_protected")
	login_protected.Use(jwtware.New(jwtware.Config{
		SigningKey: func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		},
		SigningMethod: "HS256", // Add this to specify the algorithm
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if err.Error() == "Missing or malformed JWT" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Missing or malformed JWT", "data": nil})
			}

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Unauthorized", "data": nil})
		},
	}))
	login_protected.Use(jwtBlacklist)

	edit_permission := login_protected.Group("/edit_permission")
	edit_permission.Use(NotePermissionCheck)

	login_protected.Post("/create_note", CreateGeoNote)
	api.Get("/get_by_id/:id", GetNoteByID)
	edit_permission.Delete("/edit_permission/delete_by_id/:id", DeleteNoteByID)
	edit_permission.Patch("/protected/update_by_id/:id", UpdateNoteByID)
	api.Post("/login", LogIn)
	api.Post("/register_user", RegisterUser)
	login_protected.Post("/logout", LogOut)
	log.Fatal(app.Listen(":8080"))
}

func RegisterUser(c *fiber.Ctx) error{
	var data struct{
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&data); err != nil{
		return c.Status(400).JSON(fiber.Map{"error":"Invalid request"})
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error hashing password"})
	}
	_, err = notes_db.Exec("INSERT INTO users (login, password_hash) VALUES (?, ?)", data.Username, string(hash))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "User already exists or DB error"})
	}
	return c.JSON(fiber.Map{"message": "User registered successfully"})
}

func LogIn(c *fiber.Ctx) error{
	var data struct{
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&data); err != nil{
		return c.Status(400).JSON(fiber.Map{"error":"Invalid request"})
	}
	var storedHash string
	row := notes_db.QueryRow("SELECT password_hash FROM users WHERE login=?", data.Username)
	if err := row.Scan(&storedHash); err != nil{
		return c.Status(400).JSON(fiber.Map{"error":"Incorrect username"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(data.Password)); err != nil{
		return c.Status(401).JSON(fiber.Map{"error":"Invalid password"})
	}
	claims := jwt.MapClaims{
		"name":data.Username,
		"exp":time.Now().Add(time.Hour*1).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString(jwtSecret)
	if err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"Error signing JWT"})
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"token":t})
}

func LogOut(c *fiber.Ctx) error{
	authHeader := c.Get("Authorization")
	if len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Authorization header format"})
	}
	tokenString := authHeader[7:]
	blacklist[tokenString] = true
	return c.JSON(fiber.Map{"messsage":"Logged out"})
}

func jwtBlacklist(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing Authorization header"})
	}

	tokenString := ""
	_, err := fmt.Sscan(authHeader, &tokenString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Authorization header format.  Expected 'Bearer <token>'"})
	}

	// Check if the token is blacklisted
	if blacklist[tokenString] {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Token has been blacklisted",
		})
	}

	return c.Next()
}

func CreateGeoNote(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	username := claims["name"].(string)
	var data struct{
		Text      string  `json:"text"`
		Longitude float64 `json:"longitude"`
		Lattitude float64 `json:"lattitude"`
	}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error":"Invalid request"})
	}
	_, err := notes_db.Exec("INSERT INTO notes (note_text, longitude, lattitude, user_id) VALUES (?, ?, ?, ?)", data.Text, data.Longitude, data.Lattitude, username)
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

func NotePermissionCheck(c *fiber.Ctx) error{
	id_str := c.Params("id")
	id, err := strconv.Atoi(id_str)
	if err != nil{
		return c.Status(400).JSON(fiber.Map{"error":"Invalid request"})
	}

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	username := claims["name"].(string)

	row_note_id := notes_db.QueryRow("SELECT user_id FROM notes WHERE id = ?", id)
	var user_id_note string
	row_note_id.Scan(&user_id_note)

	if username != user_id_note{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"Note does not belong to user"})
	}

	return c.Next()
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
	id := c.Params("id")
	var data NoteData
	err := c.BodyParser(&data)
	if err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"wrong json"})
	}
	_, err = notes_db.Exec("UPDATE notes SET note_text=?, longitude=?, lattitude=? WHERE id=?", data.Text, data.Longitude, data.Lattitude, id)
	if err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":fmt.Sprintf("%v", err)})
	}
	c.Status(fiber.StatusAccepted)
	return c.JSON(fiber.Map{"message":"successfully updated table"})
}