package seeders

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"joodpakpaw/pkg/storage"
	"joodpakpaw/src/models"
)

// Image content types accepted from assets/prop_mock (other files such as .html are skipped)
var propertyImageContentTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

func SeedPropertyImages(db *gorm.DB) error {
	if storage.FileStorage == nil {
		log.Println("⚠ SeedPropertyImages: FileStorage not initialized, skipping")
		return nil
	}

	mockDir := filepath.Join("database", "seeders", "assets", "prop_mock")
	ctx := context.Background()

	for _, p := range propertySeeds {
		var property models.Property
		if err := db.Where("title = ?", p.Title).First(&property).Error; err != nil {
			log.Printf("Warning: Property '%s' not found, skipping images: %v", p.Title, err)
			continue
		}

		var existingCount int64
		db.Model(&models.PropertyImage{}).Where("property_id = ?", property.ID).Count(&existingCount)
		if existingCount > 0 {
			log.Printf("- Property '%s' already has images, skipping", property.Title)
			continue
		}

		localFiles, err := listPropertyImageFiles(filepath.Join(mockDir, p.ImageDir))
		if err != nil {
			log.Printf("Warning: Failed to read image folder '%s' for '%s': %v", p.ImageDir, property.Title, err)
			continue
		}

		var images []models.PropertyImage
		for i, localPath := range localFiles {
			ext := strings.ToLower(filepath.Ext(localPath))
			fileName := fmt.Sprintf("%d_%d_%d%s", property.ID, time.Now().UnixNano(), i, ext)

			info, err := os.Stat(localPath)
			if err != nil {
				log.Printf("Warning: Failed to stat '%s': %v", localPath, err)
				continue
			}

			localFile, err := os.Open(localPath)
			if err != nil {
				log.Printf("Warning: Failed to open '%s': %v", localPath, err)
				continue
			}
			uploadErr := storage.FileStorage.Put(ctx, "properties", fileName, localFile)
			localFile.Close()

			if uploadErr != nil {
				log.Printf("Warning: Failed to upload image for '%s': %v", property.Title, uploadErr)
				continue
			}

			publicURL := storage.FileStorage.GetPublicURL("properties", fileName)
			propertyID := property.ID

			images = append(images, models.PropertyImage{
				FileName:          fileName,
				FilePath:          publicURL,
				FileContentType:   propertyImageContentTypes[ext],
				FileExtensionName: strings.TrimPrefix(ext, "."),
				FileSize:          int(info.Size()),
				PropertyID:        &propertyID,
			})
		}

		if len(images) > 0 {
			if err := db.Create(&images).Error; err != nil {
				log.Printf("Warning: Failed to save image records for '%s': %v", property.Title, err)
				continue
			}
			log.Printf("✓ Uploaded %d images for '%s'", len(images), property.Title)
		}
	}

	log.Println("✓ Seeding property images completed")
	return nil
}

// listPropertyImageFiles returns image files directly inside dir, sorted by name
func listPropertyImageFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if _, ok := propertyImageContentTypes[strings.ToLower(filepath.Ext(e.Name()))]; ok {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}
