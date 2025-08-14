package main

import (
	"log"
	"viduploader/backend/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main(){
	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://127.0.0.1:8080",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowCredentials: true,
	}))
	app.Post("/upload", handlers.UploadVideo)
	app.Static("/", "../frontend/dist")
	app.Static("/videos", "./uploads")
	log.Fatal(app.Listen(":8080"))
}

func authMiddleware(ctx *fiber.Ctx) error{
	return ctx.Next()
}