package file_module

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (m module) ServeFile(c *fiber.Ctx) error {
	// Get the file path from URL params
	// URL format: /api/v1/files/{bucket}/{path...}
	// e.g. /api/v1/files/joodpakpaw/properties/1_xxx.jpg
	fullPath := c.Params("*")
	if fullPath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file path is required"})
	}

	// Strip the bucket name prefix (first segment)
	// "joodpakpaw/properties/1_xxx.jpg" -> "properties/1_xxx.jpg"
	parts := strings.SplitN(fullPath, "/", 2)
	if len(parts) < 2 || parts[1] == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid file path"})
	}
	filePath := parts[1]

	// Sanitize path to prevent directory traversal
	filePath = filepath.Clean(filePath)
	if strings.Contains(filePath, "..") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid file path"})
	}

	// Get file from storage (MinIO)
	file, err := m.fileSystem.Get(c.Context(), filePath)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "file not found"})
	}
	tempPath := file.Name()

	// Read all bytes from file
	data, err := os.ReadFile(tempPath)
	file.Close()
	os.Remove(tempPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read file"})
	}

	// Detect content type from file bytes
	contentType := http.DetectContentType(data)

	// Set headers and send
	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=86400")

	return c.Send(data)
}
