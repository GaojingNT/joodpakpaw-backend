package seeders

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
	"joodpakpaw/pkg/storage"
	"joodpakpaw/src/models"
)

type breedEntry struct {
	Breed     string
	PetID     *int
	AssetFile string
}

func SeedPetBreeds(db *gorm.DB) error {
	dogID := 1
	catID := 2

	allBreeds := []breedEntry{
		// Dogs
		{"Golden Retriever", &dogID, "Golden_ Retriever.jpg"},
		{"Labrador Retriever", &dogID, "Labrador_Retriever.jpeg"},
		{"German Shepherd", &dogID, "German_Shepherd.jpg"},
		{"Bulldog", &dogID, "Bulldog.jpg"},
		{"Poodle", &dogID, "Poodle.jpg"},
		{"Beagle", &dogID, "Beagle.jpg"},
		{"Chihuahua", &dogID, "Chihuahua.jpg"},
		{"Shiba Inu", &dogID, "Shiba_Inu.jpg"},
		{"Siberian Husky", &dogID, "Siberian Husky.webp"},
		{"Pomeranian", &dogID, "Pomeranian.webp"},
		{"Dachshund", &dogID, "Dachshund.jpeg"},
		{"Maltese", &dogID, "Maltese.jpg"},
		{"Corgi", &dogID, "Corgi.jpg"},
		{"Border Collie", &dogID, "Border_Collie.jpg"},
		{"Rottweiler", &dogID, "Rottweiler.jpg"},
		{"Doberman", &dogID, "Doberman.jpg"},
		{"Great Dane", &dogID, "Great_Dane.jpg"},
		{"Samoyed", &dogID, "Samoyed.jpg"},
		{"Akita", &dogID, "Akita.webp"},
		{"French Bulldog", &dogID, "French_Bulldog.jpg"},
		// Cats
		{"Persian", &catID, "Persian.jpg"},
		{"Maine Coon", &catID, "Maine_Coon.jpg"},
		{"Siamese", &catID, "Siamese.avif"},
		{"British Shorthair", &catID, "British_Shorthair.avif"},
		{"Ragdoll", &catID, "Ragdoll.avif"},
		{"Bengal", &catID, "Bengal.jpg"},
		{"Abyssinian", &catID, "Abyssinian.avif"},
		{"Scottish Fold", &catID, "Scottish_Fold.webp"},
		{"Sphynx", &catID, "Sphynx.jpg"},
		{"Russian Blue", &catID, "Russian_Blue.jpg"},
		{"American Shorthair", &catID, "American_Shorthair.avif"},
		{"Norwegian Forest Cat", &catID, "Norwegian_Forest_Cat.webp"},
		{"Burmese", &catID, "Burmese.png"},
		{"Birman", &catID, "Birman.jpg"},
		{"Tonkinese", &catID, "Tonkinese.webp"},
		{"Manx", &catID, "Manx.jpg"},
		{"Turkish Angora", &catID, "Turkish_Angor.jpg"},
		{"Devon Rex", &catID, "Devon_Rex.jpg"},
		{"Cornish Rex", &catID, "Cornish_Rex.jpg"},
		{"Himalayan", &catID, "Himalayan.jpg"},
	}

	ctx := context.Background()
	assetsDir := filepath.Join("database", "seeders", "assets")

	for _, entry := range allBreeds {
		var existing models.PetBreed
		result := db.Where("breed = ? AND pet_id = ?", entry.Breed, entry.PetID).First(&existing)

		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			log.Printf("Error checking breed '%s': %v", entry.Breed, result.Error)
			return result.Error
		}

		if result.Error == gorm.ErrRecordNotFound {
			imageURL := uploadBreedImage(ctx, assetsDir, entry.AssetFile, entry.Breed)
			breed := models.PetBreed{
				Breed:    entry.Breed,
				PetID:    entry.PetID,
				ImageURL: imageURL,
			}
			if err := db.Create(&breed).Error; err != nil {
				log.Printf("Failed to create breed '%s': %v", entry.Breed, err)
				return err
			}
			log.Printf("✓ Pet breed '%s' created successfully", entry.Breed)
		} else if existing.ImageURL == nil {
			imageURL := uploadBreedImage(ctx, assetsDir, entry.AssetFile, entry.Breed)
			if imageURL != nil {
				if err := db.Model(&existing).Update("image_url", imageURL).Error; err != nil {
					log.Printf("Warning: Failed to update image for breed '%s': %v", entry.Breed, err)
				} else {
					log.Printf("✓ Updated image for breed '%s'", entry.Breed)
				}
			} else {
				log.Printf("- Pet breed '%s' already exists (no image), skipping", entry.Breed)
			}
		} else {
			log.Printf("- Pet breed '%s' already exists, skipping", entry.Breed)
		}
	}

	return nil
}

func uploadBreedImage(ctx context.Context, assetsDir, assetFile, breedName string) *string {
	if storage.FileStorage == nil {
		log.Printf("⚠ FileStorage not initialized, skipping image upload for '%s'", breedName)
		return nil
	}

	localPath := filepath.Join(assetsDir, assetFile)
	file, err := os.Open(localPath)
	if err != nil {
		log.Printf("Warning: Could not open asset '%s': %v", assetFile, err)
		return nil
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(assetFile))
	safeName := strings.ReplaceAll(breedName, " ", "_")
	fileName := fmt.Sprintf("%s%s", safeName, ext)

	if err := storage.FileStorage.Put(ctx, "pet-breeds", fileName, file); err != nil {
		log.Printf("Warning: Failed to upload image for breed '%s': %v", breedName, err)
		return nil
	}

	url := storage.FileStorage.GetPublicURL("pet-breeds", fileName)
	return &url
}
