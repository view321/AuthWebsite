package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
	"math/rand/v2"
	"net/smtp"
	"strings"

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
	ParentID     int      `json:"parent_id"`
}

var notes_db *sql.DB
var smtp_client *smtp.Client
var jwtSecret = []byte("my_key_123")

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

	smtp_client, err = setupSMTPClient()
	if err != nil {
		log.Fatal(err)
	}
	defer smtp_client.Quit() 

	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowMethods:     "POST, GET, DELETE, PATCH, OPTIONS",
		AllowOrigins:     "http://127.0.0.1:8080, http://localhost:8080",
		AllowHeaders:     "Origin, Content-Type, Accept, Authentication, Set-Cookie, Cookie",
		AllowCredentials: true,
	}))
	api := app.Group("/api")

	login_protected := api.Group("/login_protected")
	login_protected.Use(jwtMiddleware(jwtSecret))

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
	api.Post("/send_email_key", SendEmailKey)
	api.Post("/register_user", RegisterUser)

	login_protected.Post("/logout", LogOut)
	login_protected.Get("/check_session", CheckSession)

	app.Static("/", "../frontend/")
	log.Fatal(app.Listen(":8080"))
}

func optionalJWT(secretKey []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		jwt_from_cookie := c.Cookies("jwt")
		tokenString := jwt_from_cookie

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
			if !jwtBlacklisted(tokenString) {
				c.Locals("Registered", true)
				c.Locals("user", claims)
				return c.Next()
			}
		}

		c.Locals("Registered", false)
		return c.Next()
	}
}

func jwtBlacklisted(jwt string) bool {
	row := notes_db.QueryRow("SELECT jwt FROM blacklist WHERE jwt=?", jwt)
	var __jwt string
	err := row.Scan(&__jwt)
	return err == nil
}

func blacklist_jwt(jwt string) {
	notes_db.Exec("INSERT INTO blacklist jwt VALUES ?", jwt)
}
func jwtMiddleware(secretKey []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		jwt_from_cookie := c.Cookies("jwt")
		if jwt_from_cookie == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No token provided"})
		}

		if jwtBlacklisted(jwt_from_cookie) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token is blacklisted"})
		}

		token, err := jwt.Parse(jwt_from_cookie, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil
		})

		if err != nil {
			if err.Error() == "token is expired" {
				// Clear the expired cookie
				c.Cookie(&fiber.Cookie{
					Name:     "jwt",
					Value:    "",
					Expires:  time.Now().Add(-24 * time.Hour),
					Path:     "/",
					HTTPOnly: true,
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token expired"})
				}
			}
			c.Locals("user", claims)
			return c.Next()
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}
}

func SendEmailKey(c *fiber.Ctx) error{
	var email struct {
		Email string `json:"email"`
	}
	err := c.BodyParser(&email.Email)
	if err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"Cannot parse error"})
	}
	randomNumber := rand.IntN(900000) + 100000

	row := notes_db.QueryRow("SELECT email_key FROM email_keys WHERE email=?", email.Email)
	var test_for_scan string
	err = row.Scan(&test_for_scan)
	if err == nil{
		_, err := notes_db.Exec("UPDATE email_keys SET email_key=? WHERE email=?", strconv.Itoa(randomNumber), email.Email)
		if err != nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"Cannot update table"})
		}
	}else{
		_, err := notes_db.Exec("INSERT INTO email_keys email, email_key VALUES ?, ?", email.Email, strconv.Itoa(randomNumber))
		if err != nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"Cannot add row to table"})
		}
	}

	from := "verify_email@sandbox2c950af4cb4743ccb82a5a406e062642.mailgun.org"
    to := []string{email.Email}

    subject := "Verification code"
    body := strconv.Itoa(randomNumber)
	msg := "From: " + from + "\r\n" +
        "To: " + strings.Join(to, ",") + "\r\n" +
        "Subject: " + subject + "\r\n" +
        "\r\n" +
        body + "\r\n"
	if err := smtp_client.Mail(from); err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"Cannot start sending email"})
	}
	for _, address := range to {
        if err := smtp_client.Rcpt(address); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"Issue with rcpt"})
        }
    }
	w, err := smtp_client.Data()
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"Issue with smtp_client.Data()"})
    }
	_, err = w.Write([]byte(msg))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"Issue with writing to client"})
    }
    err = w.Close()
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"Issue with closing client"})
    }
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message":"Email sent successfully"})
}

func RegisterUser(c *fiber.Ctx) error {
	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email string `json:"email"`
		EmailKey    string `json:"email_key"`
	}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	row := notes_db.QueryRow("SELECT email_key FROM email_keys WHERE email=?", data.Email)
	var email_key string
	err  := row.Scan(&email_key)
	if err != nil{
		return c.Status(400).JSON(fiber.Map{"error": "Cannot scan email key"})
	}
	if email_key != data.EmailKey{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"wrong email key"})
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error hashing password"})
	}
	_, err = notes_db.Exec("INSERT INTO users (login, password_hash) VALUES (?, ?)", data.Username, string(hash))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "User already exists or DB error"})
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
	cookie := new(fiber.Cookie)
	cookie.Name = "jwt"
	cookie.Value = t
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.Path = "/"      // Add this
	cookie.HTTPOnly = true // Add this
	c.Cookie(cookie)
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "user registered  and logged in successfully", "token": t})
}
func LogIn(c *fiber.Ctx) error {
	var username string

	jwt_from_cookie := c.Cookies("jwt")

	isBlacklisted := jwtBlacklisted(jwt_from_cookie)

	if jwt_from_cookie != "" && !isBlacklisted {
		token, err := jwt.Parse(jwt_from_cookie, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				username, _ = claims["name"].(string)
			}
		}
	}

	if username == "" {
		var data struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&data); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication failed"})
		}

		var storedHash string
		row := notes_db.QueryRow("SELECT password_hash FROM users WHERE login=?", data.Username)
		if err := row.Scan(&storedHash); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication failed"})
		}

		if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(data.Password)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication failed"})
		}
		username = data.Username
	}

	if username != "" {
		if jwt_from_cookie != "" {

		}
		expirationTime := time.Now().Add(24 * time.Hour)
		claims := jwt.MapClaims{
			"name": username,
			"exp":  expirationTime.Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		t, err := token.SignedString(jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not generate token",
			})
		}

		cookie := &fiber.Cookie{
			Name:     "jwt",
			Value:    t,
			Path:     "/",
			Expires:  expirationTime,
			HTTPOnly: true,
			SameSite: "Lax",
		}
		c.Cookie(cookie)

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"token": t,
			"user":  username,
		})
	}

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": "Authentication failed",
	})
}
func LogOut(c *fiber.Ctx) error {
	jwt_from_cookie := c.Cookies("jwt")

	cookie := &fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		Domain:   "",
		Expires:  time.Now().Add(-24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   false,
	}
	c.Cookie(cookie)

	c.Cookie(&fiber.Cookie{
		Name:  "jwt",
		Value: "",
		Path:  "",
	})

	if jwt_from_cookie != "" {
		blacklist_jwt(jwt_from_cookie)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

func CreateGeoNote(c *fiber.Ctx) error {
	claims := c.Locals("user").(jwt.MapClaims)
	username, ok := claims["name"].(string)
	if !ok {
		fmt.Println("Invalid username in claims")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid username in claims"})
	}
	var data struct {
		Text         string   `json:"text"`
		Longitude    float64  `json:"longitude"`
		Lattitude    float64  `json:"lattitude"`
		Public       bool     `json:"public"`
		AllowedUsers []string `json:"allowed_users"`
		ParentID     int      `json:"parent_id"`
	}
	if err := c.BodyParser(&data); err != nil {
		fmt.Println("Invalid request")
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	geo_identical_note := notes_db.QueryRow("SELECT note_text FROM notes WHERE longitude=? AND lattitude=?", data.Longitude, data.Lattitude)
	var identical_test_text string
	if err := geo_identical_note.Scan(&identical_test_text); err == nil {
		fmt.Println("There is already a note in the position")
		return c.Status(400).JSON(fiber.Map{"error": "Invalid position for note"})
	}

	var res sql.Result
	if data.ParentID == 0 {
		var err error
		res, err = notes_db.Exec("INSERT INTO notes (note_text, longitude, lattitude, user_id, public, parent_id) VALUES (?, ?, ?, ?, ?, ?)", data.Text, data.Longitude, data.Lattitude, username, data.Public, data.ParentID)
		if err != nil {
			fmt.Println(err)
			return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("%v", err)})
		}
	} else {
		row := notes_db.QueryRow("SELECT public, longitude, lattitude FROM notes WHERE id=?", data.ParentID)
		var is_public bool
		var longitude, lattitude float64
		row.Scan(&is_public, &longitude, &lattitude)
		if is_public {
			_, err := notes_db.Exec("INSERT INTO notes (note_text, longitude, lattitude, user_id, public, parent_id) VALUES (?, ?, ?, ?, ?, ?)", data.Text, longitude, lattitude, username, true, data.ParentID)
			if err != nil {
				fmt.Println(err)
				return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("%v", err)})
			}
			return c.JSON(fiber.Map{"message": "Note added successfully"})
		} else {
			rows, err := notes_db.Query("SELECT user_login FROM note_access WHERE note_id=?", data.ParentID)
			if err != nil {
				fmt.Println(err)
				return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("%v", err)})
			}
			res, err = notes_db.Exec("INSERT INTO notes (note_text, longitude, lattitude, user_id, public, parent_id) VALUES (?, ?, ?, ?, ?, ?)", data.Text, longitude, lattitude, username, false, data.ParentID)
			if err != nil {
				fmt.Println(err)
				return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("%v", err)})
			}
			note_id, err := res.LastInsertId()
			if err != nil {
				fmt.Println(err)
				return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("%v", err)})
			}
			for rows.Next() {
				var new_user string
				rows.Scan(&new_user)
				_, err := notes_db.Exec("INSERT note_id, user_login INTO note_access VALUES ?, ?", note_id, new_user)
				if err != nil {
					fmt.Println(err)
					return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("%v", err)})
				}
			}
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Note added successfully"})
		}
	}

	if !data.Public {
		note_id, err := res.LastInsertId()
		if err != nil {
			fmt.Println("LastInsertId did not work")
			return c.Status(400).JSON(fiber.Map{"error": "Cannot extract last inserted id"})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad json"})
	}

	rows, err := notes_db.Query(`SELECT id, note_text, longitude, lattitude, user_id, public, parent_id FROM notes 
	WHERE longitude > ? AND longitude < ? AND lattitude > ? AND lattitude < ?`, data.Lower_longitude,
		data.Upper_longitude, data.Lower_lattitude, data.Upper_lattitude)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Bad database request"})
	}
	defer rows.Close()

	var notes []NoteData
	for rows.Next() {
		var note NoteData
		err := rows.Scan(&note.ID, &note.Text, &note.Longitude, &note.Lattitude, &note.UserID, &note.Public, &note.ParentID)
		if err != nil {
			fmt.Printf("Scan error: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Error scanning note data: %v", err)})
		}
		availiable, err := CheckIfAvailiable(c, note.ID)
		if err != nil {
			fmt.Printf("CheckIfAvailiable error for note ID %d: %v\n", note.ID, err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("availiability check failed: %v", err)})
		}
		if availiable {
			notes = append(notes, note)
		}
	}

	if err = rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error iterating over notes"})
	}

	c.Status(fiber.StatusAccepted)
	return c.JSON(notes)
}

func NotePermissionCheck(c *fiber.Ctx) error {
	userClaims := c.Locals("user")
	if userClaims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not authenticated"})
	}
	claims := userClaims.(jwt.MapClaims)
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
	err = row_note_id.Scan(&user_id_note)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Note not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	if username != user_id_note {
		fmt.Println(username + " " + user_id_note)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Note does not belong to user or does not exist"})
	}
	c.Locals("id", id)
	return c.Next()
}

func DeleteNoteByID(c *fiber.Ctx) error {
	id := c.Locals("id")
	child_row := notes_db.QueryRow("SELECT note_text FROM notes WHERE parent_id = ?", id)
	var note_text string
	if err := child_row.Scan(&note_text); err == nil {
		_, err := notes_db.Exec("UPDATE notes SET note_text = \"[Deleted]\" WHERE id=?", id)
		if err != nil {
			fmt.Println("Error clearing a node")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error clearing a node"})
		}
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Successfully cleared note"})
	}
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
	if data.ParentID != 0 {
		var parent_data NoteData
		row := notes_db.QueryRow("SELECT longitude, lattitude FROM notes WHERE id=?", data.ParentID)
		err := row.Scan(&parent_data.Longitude, &parent_data.Lattitude)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot parse parent"})
		}
		data.Lattitude = parent_data.Lattitude
		data.Longitude = parent_data.Longitude
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Bad database request"})
	}
	defer rows.Close()

	var notes []NoteData
	for rows.Next() {
		var note NoteData
		err := rows.Scan(&note.ID, &note.Text, &note.Longitude, &note.Lattitude, &note.UserID, &note.Public)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error scanning note data"})
		}
		availiable, err := CheckIfAvailiable(c, note.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error checking availiability"})
		}
		if availiable {
			notes = append(notes, note)
		}
	}

	if err = rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error iterating over notes"})
	}

	c.Status(fiber.StatusAccepted)
	return c.JSON(notes)
}

func CheckIfAvailiable(c *fiber.Ctx, id int) (bool, error) {
	public_row := notes_db.QueryRow("SELECT public FROM notes WHERE id=?", id)
	var public bool
	err := public_row.Scan(&public)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("note not found")
		}
		return false, err
	}

	if public {
		return true, nil
	} else {
		// Check if user is registered - handle nil case
		registeredValue := c.Locals("Registered")
		if registeredValue == nil {
			return false, errors.New("registration status not found")
		}
		registered, ok := registeredValue.(bool)
		if !ok {
			return false, errors.New("invalid registration status type")
		}
		if !registered {
			return false, errors.New("not registered")
		}

		claims := c.Locals("user").(jwt.MapClaims)
		username, ok := claims["name"].(string)
		if !ok {
			return false, errors.New("invalid username in claims")
		}

		// First check if the user is the owner of the note
		owner_row := notes_db.QueryRow("SELECT user_id FROM notes WHERE id=?", id)
		var note_owner string
		err := owner_row.Scan(&note_owner)
		if err != nil {
			return false, errors.New("note not found")
		}
		if note_owner == username {
			return true, nil
		}

		// Then check if the user has access through the note_access table
		pub_rows, err := notes_db.Query("SELECT user_login FROM note_access WHERE note_id=?", id)
		if err != nil {
			return false, errors.New("invalid database request")
		}
		defer pub_rows.Close()

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

func CheckSession(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not parse user claims from token"})
	}
	username, ok := claims["name"].(string)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Username not found in token"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "session is valid",
		"user":    username,
	})
}

func setupSMTPClient() (*smtp.Client, error) {
    auth := smtp.PlainAuth("", "verify_email@sandbox2c950af4cb4743ccb82a5a406e062642.mailgun.org", "C7fX23VR9n@5t!@", "smtp.mailgun.org")
    client, err := smtp.Dial("smtp.mailgun.org:587")
    if err != nil {
        return nil, err
    }
    if err := client.Auth(auth); err != nil {
        return nil, err
    }
    return client, nil
}