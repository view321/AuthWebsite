package handlers

import (
	"fmt"
 	"io"
 	"log"
 	"mime/multipart"
 	"net/http"
 	"os"
 	"path/filepath"
 	"strings"
 	"time"

 	"github.com/gofiber/fiber/v2"
 	"github.com/google/uuid" 
)

func UploadVideo(c *fiber.Ctx) error {
 	// Get the video file from the request
 	file, err := c.FormFile("video")
 	if err != nil {
 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to retrieve video file"})
 	}

 	// Validate file type (important for security)
 	if !isValidVideoType(file.Header["Content-Type"][0]) {
 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid video file type"})
 	}

 	// Generate a unique filename
 	filename := generateUniqueFilename(file.Filename)

 	// Define the destination path (where to store the video)
 	destination := filepath.Join("./uploads", filename) //  Creates the "uploads" directory if it doesn't exist

 	// Create the "uploads" directory if it doesn't exist
 	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
 		os.Mkdir("./uploads", 0755) // Creates the directory with read/write/execute permissions for the owner, and read/execute permissions for others.
 	}

 	// Save the file to the destination
 	if err := c.SaveFile(file, destination); err != nil {
 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save video file"})
 	}

 	// Return success message
 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
 		"message":  "Video uploaded successfully",
 		"filename": filename, // Return the filename for later use
 		"url":      "/videos/" + filename, // Example URL to access the video (you might need to serve files statically)
 	})
 }

 func generateUniqueFilename(originalFilename string) string {
 	ext := filepath.Ext(originalFilename)
 	name := strings.TrimSuffix(originalFilename, ext)
 	uniqueID := uuid.New().String()
 	timestamp := time.Now().Format("20060102150405") //YYYYMMDDHHMMSS
 	return fmt.Sprintf("%s_%s_%s%s", name, timestamp, uniqueID, ext) // avoid duplicate names
 }

func isValidVideoType(contentType string) bool {
 	allowedTypes := []string{"video/mp4", "video/webm", "video/ogg", "video/quicktime"} // Common video types
 	for _, t := range allowedTypes {
 		if contentType == t {
 			return true
 		}
 	}
 	return false
 }