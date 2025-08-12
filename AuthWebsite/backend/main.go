package main

import (
    "database/sql"
    "fmt"
    "log"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/cors"
    _ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var sessions = make(map[string]string) // sessionToken → username

func main() {
    var err error

    // Connect to MySQL
    dsn := "root:347347@tcp(127.0.0.1:3306)/passwords_db"
    db, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }
    if err = db.Ping(); err != nil {
        log.Fatal(err)
    }

    // Create users table if not exists
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INT AUTO_INCREMENT PRIMARY KEY,
            username VARCHAR(255) UNIQUE NOT NULL,
            password VARCHAR(255) NOT NULL
        )
    `)
    if err != nil {
        log.Fatal(err)
    }

    // Fiber app
    app := fiber.New()
    app.Use(logger.New())
    app.Use(cors.New(cors.Config{
        AllowOrigins: "http://127.0.0.1:8080",
        AllowHeaders: "Origin, Content-Type, Accept",
    }))

    // API routes
    api := app.Group("/api")
    api.Post("/register", registerHandler)
    api.Post("/login", loginHandler)

    protected := api.Group("/", authMiddleware)
    protected.Get("/profile", profileHandler)
    protected.Post("/logout", logoutHandler)

    // Serve static files from Svelte build
    app.Static("/", "../frontend/dist")

    log.Fatal(app.Listen(":8080"))
}

func registerHandler(c *fiber.Ctx) error {
    var data struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    if err := c.BodyParser(&data); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }

    _, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", data.Username, data.Password)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "User already exists or DB error"})
    }

    return c.JSON(fiber.Map{"message": "User registered successfully"})
}

func loginHandler(c *fiber.Ctx) error {
    var data struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    if err := c.BodyParser(&data); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }

    var storedPassword string
    err := db.QueryRow("SELECT password_hash FROM users WHERE username = ?", data.Username).Scan(&storedPassword)
    if err != nil || storedPassword != data.Password {
        return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
    }

    // Create a session token
    token := fmt.Sprintf("%d", time.Now().UnixNano())
    sessions[token] = data.Username

    // Set cookie
    c.Cookie(&fiber.Cookie{
        Name:     "session_token",
        Value:    token,
        Expires:  time.Now().Add(24 * time.Hour),
        HTTPOnly: true,
    })

    return c.JSON(fiber.Map{"message": "Login successful"})
}

func authMiddleware(c *fiber.Ctx) error {
    token := c.Cookies("session_token")
    username, exists := sessions[token]
    if !exists {
        return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
    }

    // Store username in context
    c.Locals("username", username)

    return c.Next()
}

func profileHandler(c *fiber.Ctx) error {
    username := c.Locals("username").(string)
    return c.JSON(fiber.Map{
        "message":  "Welcome to your profile",
        "username": username,
    })
}

func logoutHandler(c *fiber.Ctx) error {
    token := c.Cookies("session_token")
    delete(sessions, token)

    // Clear cookie
    c.Cookie(&fiber.Cookie{
        Name:     "session_token",
        Value:    "",
        Expires:  time.Now().Add(-1 * time.Hour),
        HTTPOnly: true,
    })

    return c.JSON(fiber.Map{"message": "Logged out successfully"})
}