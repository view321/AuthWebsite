package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type NoteData struct {
	ID           int      `json:"id"`
	Text         string   `json:"text"`
	Longitude    float64  `json:"longitude"`
	Lattitude    float64  `json:"lattitude"`
	Public       bool     `json:"public"`
	UserID       string   `json:"user_id"`
	AllowedUsers []string `json:"allowed_users"`
}

var notes_db *sql.DB
var jwtSecret = []byte("my_key_123")
var blacklist = make(map[string]bool)

func main() {
	dsn := "root:347347@tcp(127.0.0.1:3306)/notes_db"
	var err error
	notes_db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer notes_db.Close()
	if err = notes_db.Ping(); err != nil {
		log.Fatal(err)
	}
	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowMethods:     "*",
		AllowOrigins:     "http://127.0.0.1:8080",
		AllowHeaders:     "Origin, Content-Type, Accept, Authentication",
		AllowCredentials: true,
	}))
	api := app.Group("/api")

	login_protected := api.Group("/login_protected")
	login_protected.Use(jwtMiddleware(jwtSecret))
	login_protected.Use(jwtBlacklist)

	edit_permission := login_protected.Group("/edit_permission/:id")
	edit_permission.Use(NotePermissionCheck)

	view_permission := api.Group("/view_permission")
	view_permission.Use(optionalJWT(jwtSecret))

	login_protected.Post("/create_note", CreateGeoNote)

	view_permission.Get("/get_by_id/:id", GetNoteByID)
	view_permission.Get("/get_by_user/:user_id", GetNotesByUser)
	view_permission.Post("/get_within_square", GetNotesWithinSquare)

	edit_permission.Delete("/delete", DeleteNoteByID)
	edit_permission.Patch("/update", UpdateNoteByID)
	api.Post("/login", LogIn)
	api.Post("/register_user", RegisterUser)
	login_protected.Post("/logout", LogOut)

	app.Static("/", "../frontend/")
	log.Fatal(app.Listen(":8080"))
}

func optionalJWT(secretKey []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			c.Locals("Registered", false)
			return c.Next()
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Locals("Registered", false)
			return c.Next()
		}
		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Check the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil
		})

		if err != nil {
			c.Locals("Registered", false)
			return c.Next()
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if blacklist[tokenString] {
				c.Locals("Registered", false)
				return c.Next()
			}
			c.Locals("Registered", true)
			c.Locals("user", claims)
			return c.Next()
		}

		c.Locals("Registered", false)
		return c.Next()
	}
}

func jwtMiddleware(secretKey []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing Authorization header"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Authorization header format. Expected 'Bearer <token>'"})
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Check the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil
		})

		if err != nil {
			log.Printf("JWT Parsing Error: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid JWT token"})
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Store the claims in the context for later use
			c.Locals("user", claims) // or c.Locals("claims", claims)
			return c.Next()
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid JWT token"})
	}
}

func RegisterUser(c *fiber.Ctx) error {
	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
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

func LogIn(c *fiber.Ctx) error {
	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	var storedHash string
	row := notes_db.QueryRow("SELECT password_hash FROM users WHERE login=?", data.Username)
	if err := row.Scan(&storedHash); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Incorrect username"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(data.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid password"})
	}
	claims := jwt.MapClaims{
		"name": data.Username,
		"exp":  time.Now().Add(time.Hour * 1).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error signing JWT"})
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"token": t})
}

func LogOut(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Authorization header format"})
	}
	tokenString := authHeader[7:]
	blacklist[tokenString] = true
	return c.JSON(fiber.Map{"messsage": "Logged out"})
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
	claims := c.Locals("user").(jwt.MapClaims)
	username, ok := claims["name"].(string)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid username in claims"})
	}
	var data struct {
		Text         string   `json:"text"`
		Longitude    float64  `json:"longitude"`
		Lattitude    float64  `json:"lattitude"`
		Public       bool     `json:"public"`
		AllowedUsers []string `json:"allowed_users"`
	}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	res, err := notes_db.Exec("INSERT INTO notes (note_text, longitude, lattitude, user_id, public) VALUES (?, ?, ?, ?, ?)", data.Text, data.Longitude, data.Lattitude, username, data.Public)
	if err != nil {
		
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("%v", err)})
	}

	if !data.Public {
		note_id, err := res.LastInsertId()
		if err != nil {
			fmt.Println("LastInsertId did not work")
			return c.Status(400).JSON(fiber.Map{"error": "Cannot extract las inserted id"})
		}
		for _, allowed_user := range data.AllowedUsers {
			_, err := notes_db.Exec("INSERT INTO note_access (note_id, user_login) VALUES (?, ?)", int(note_id), allowed_user)
			if err != nil {
				fmt.Println(note_id)
				fmt.Println(allowed_user)
				fmt.Println(err)
				return c.Status(400).JSON(fiber.Map{"error": "Cannot paste in note_access"})
			}
		}
	}

	return c.JSON(fiber.Map{"message": "Note added successfully"})
}

func GetNoteByID(c *fiber.Ctx) error {
	id_str := c.Params("id")
	id, err := strconv.Atoi(id_str)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	row := notes_db.QueryRow("SELECT note_text, longitude, lattitude, user_id FROM notes WHERE id = ?", id)
	var data NoteData
	availiable, err := CheckIfAvailiable(c, id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("%v", err)})
	}
	if availiable {
		data.ID = id
		err = row.Scan(&data.Text, &data.Longitude, &data.Lattitude, &data.UserID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("%v", err)})
		}
		return c.JSON(data)
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "note not availiable"})
	}
}

func GetNotesWithinSquare(c *fiber.Ctx) error {
	var data struct {
		Upper_lattitude float64 `json:"upper_lattitude"`
		Lower_lattitude float64 `json:"lower_lattitude"`
		Upper_longitude float64 `json:"upper_longitude"`
		Lower_longitude float64 `json:"lower_longitude"`
	}
	err := c.BodyParser(&data)
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad json"})
	}

	rows, err := notes_db.Query(`SELECT id, note_text, longitude, lattitude, user_id, public FROM notes 
	WHERE longitude > ? AND longitude < ? AND lattitude > ? AND lattitude < ?`, data.Lower_longitude,
		data.Upper_longitude, data.Lower_lattitude, data.Upper_lattitude)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Bad database request"})
	}
	var notes []NoteData
	for rows.Next() {
		var note NoteData
		rows.Scan(&note.ID, &note.Text, &note.Longitude, &note.Lattitude, &note.UserID, &note.Public)
		availiable, err := CheckIfAvailiable(c, note.ID)
		if err != nil {
			c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "availiability check failed"})
		}
		if availiable {
			notes = append(notes, note)
		}
	}
	c.Status(fiber.StatusAccepted)
	return c.JSON(notes)
}

func NotePermissionCheck(c *fiber.Ctx) error {
	claims := c.Locals("user").(jwt.MapClaims)
	username, ok := claims["name"].(string)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid username in claims"})
	}

	id_str := c.Params("id")
	id, err := strconv.Atoi(id_str)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	row_note_id := notes_db.QueryRow("SELECT user_id FROM notes WHERE id = ?", id)
	var user_id_note string
	row_note_id.Scan(&user_id_note)

	if username != user_id_note {
		fmt.Println(username + " " + user_id_note)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Note does not belong to user or does not exist"})
	}
	c.Locals("id", id)
	return c.Next()
}

func DeleteNoteByID(c *fiber.Ctx) error {
	id := c.Locals("id")
	_, err := notes_db.Exec("DELETE FROM notes WHERE id = ?", id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid database request"})
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Successfully deleted note"})
}

func UpdateNoteByID(c *fiber.Ctx) error {
	id := c.Locals("id")
	var data NoteData
	err := c.BodyParser(&data)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "wrong json"})
	}
	_, err = notes_db.Exec("UPDATE notes SET note_text=?, longitude=?, lattitude=? WHERE id=?", data.Text, data.Longitude, data.Lattitude, id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("%v", err)})
	}
	c.Status(fiber.StatusAccepted)
	return c.JSON(fiber.Map{"message": "successfully updated table"})
}

func GetNotesByUser(c *fiber.Ctx) error {
	user_id := c.Params("user_id")
	fmt.Println(user_id)
	rows, err := notes_db.Query(`SELECT * FROM notes WHERE user_id=?`, user_id)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Bad database request"})
	}
	var notes []NoteData
	for rows.Next() {
		var note NoteData
		rows.Scan(&note.ID, &note.Text, &note.Longitude, &note.Lattitude, &note.UserID, &note.Public)
		availiable, err := CheckIfAvailiable(c, note.ID)
		if err != nil {
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error checking availiability"})
		}
		if availiable {
			notes = append(notes, note)
		}
	}
	c.Status(fiber.StatusAccepted)
	return c.JSON(notes)
}

func CheckIfAvailiable(c *fiber.Ctx, id int) (bool, error) {
	public_row := notes_db.QueryRow("SELECT public FROM notes WHERE id=?", id)
	var public bool
	public_row.Scan(&public)
	if public {
		return true, nil
	} else {
		var registered bool = c.Locals("Registered").(bool)
		if !registered {
			return false, errors.New("not registered")
		}

		claims := c.Locals("user").(jwt.MapClaims)
		username, ok := claims["name"].(string)
		if !ok {
			return false, errors.New("invalid username in claims")
		}
		pub_rows, err := notes_db.Query("SELECT user_login FROM note_access WHERE note_id=?", id)
		if err != nil {
			return false, errors.New("invalid database request")
		}
		for pub_rows.Next() {
			var login_from_access_table string
			err := pub_rows.Scan(&login_from_access_table)
			if err != nil {
				return false, err
			}
			if login_from_access_table == username {
				return true, nil
			}
		}
		return false, nil
	}
}
