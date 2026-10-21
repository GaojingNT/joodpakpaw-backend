package room_category_module

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/room_category_module/dtos"
)

func (s Service) GetAllRoomCategories(ctx context.Context) ([]models.RoomCategory, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetAllRoomCategoriesService", trace.WithAttributes(attribute.String("service", "GetAllRoomCategories")))

	roomCategories, err := s.roomCategoryRepository().GetAllRoomCategories(ctx)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	s.tracer.TraceEnd(childSpan)

	return roomCategories, nil
}

func (s Service) GetRoomCategoriesByPropertyID(ctx context.Context, propertyID int) ([]models.RoomCategory, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetRoomCategoriesByPropertyIDService", trace.WithAttributes(attribute.String("service", "GetRoomCategoriesByPropertyID"), attribute.Int64("propertyId", int64(propertyID))))

	roomCategories, err := s.roomCategoryRepository().GetRoomCategoriesByPropertyID(ctx, propertyID)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	s.tracer.TraceEnd(childSpan)

	return roomCategories, nil
}

func (s Service) GetRoomCategoryByID(ctx context.Context, id int) (models.RoomCategory, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetRoomCategoryByIDService", trace.WithAttributes(attribute.String("service", "GetRoomCategoryByID"), attribute.Int64("id", int64(id))))

	roomCategory, err := s.roomCategoryRepository().GetRoomCategoryByID(ctx, id)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return models.RoomCategory{}, err
	}

	s.tracer.TraceEnd(childSpan)

	return roomCategory, nil
}

func (s Service) CreateRoomCategory(ctx context.Context, roomCategoryDto *dtos.CreateRoomCategory) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "CreateRoomCategoryService", trace.WithAttributes(attribute.String("service", "CreateRoomCategory")))

	// Check if property already has 10 room categories
	count, err := s.roomCategoryRepository().CountRoomCategoriesByPropertyID(ctx, *roomCategoryDto.PropertyID)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}

	if count >= 10 {
		s.tracer.TraceEnd(childSpan)
		return exception.ErrMaxRoomCategoriesExceeded
	}

	roomCategory := new(models.RoomCategory)
	roomCategory.Name = roomCategoryDto.Name
	roomCategory.TotalRoom = roomCategoryDto.TotalRoom
	roomCategory.NumberOfPets = roomCategoryDto.NumberOfPets
	roomCategory.BasePrice = roomCategoryDto.BasePrice
	roomCategory.PropertyID = roomCategoryDto.PropertyID
	roomCategory.PetSizeCategoryID = roomCategoryDto.PetSizeCategoryID

	// Load pets from database if pet IDs are provided
	if len(roomCategoryDto.Pets) > 0 {
		pets, err := s.roomCategoryRepository().GetPetsByIDs(ctx, roomCategoryDto.Pets)
		if err != nil {
			s.tracer.TraceEnd(childSpan)
			return err
		}
		roomCategory.Pets = pets
	}

	// Load facilities from database if facility IDs are provided
	if len(roomCategoryDto.Facilities) > 0 {
		facilities, err := s.roomCategoryRepository().GetFacilitiesByIDs(ctx, roomCategoryDto.Facilities)
		if err != nil {
			s.tracer.TraceEnd(childSpan)
			return err
		}
		roomCategory.Facilities = facilities
	}

	// Create room category in database with calendar entries
	err = s.roomCategoryRepository().CreateRoomCategoryWithCalendar(ctx, roomCategory)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}

	// Handle image uploads if images are provided
	if len(roomCategoryDto.Images) > 0 {
		roomCategoryImages := make([]models.RoomCategoryImage, 0, len(roomCategoryDto.Images))
		basePath := "room-categories"

		for index, fileHeader := range roomCategoryDto.Images {
			timestamp := time.Now().UnixNano()
			ext := filepath.Ext(fileHeader.Filename)
			fileName := fmt.Sprintf("%d_%d_%d%s", roomCategory.ID, timestamp, index, ext)

			if err := s.fileSystem.Put(ctx, basePath, fileName, fileHeader); err != nil {
				s.logger.Error(fmt.Sprintf("Failed to upload room category image: %v", err))
				continue
			}

			publicURL := s.fileSystem.GetPublicURL(basePath, fileName)

			roomCategoryImages = append(roomCategoryImages, models.RoomCategoryImage{
				FileName:          fileHeader.Filename,
				FilePath:          publicURL,
				FileContentType:   fileHeader.Header.Get("Content-Type"),
				FileExtensionName: strings.TrimPrefix(ext, "."),
				FileSize:          int(fileHeader.Size),
				RoomCategoryID:    &roomCategory.ID,
			})
		}

		if len(roomCategoryImages) > 0 {
			if err = s.roomCategoryRepository().CreateRoomCategoryImages(ctx, roomCategoryImages); err != nil {
				s.logger.Error(fmt.Sprintf("Failed to save room category images: %v", err))
			}
		}
	}

	s.tracer.TraceEnd(childSpan)

	return nil
}

func (s Service) UpdateRoomCategory(ctx context.Context, id int, roomCategoryDto *dtos.UpdateRoomCategory) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UpdateRoomCategoryService", trace.WithAttributes(attribute.String("service", "UpdateRoomCategory"), attribute.Int64("id", int64(id))))

	roomCategory := new(models.RoomCategory)
	roomCategory.Name = roomCategoryDto.Name
	roomCategory.TotalRoom = roomCategoryDto.TotalRoom
	roomCategory.NumberOfPets = roomCategoryDto.NumberOfPets
	roomCategory.BasePrice = roomCategoryDto.BasePrice
	roomCategory.PropertyID = roomCategoryDto.PropertyID
	roomCategory.PetSizeCategoryID = roomCategoryDto.PetSizeCategoryID

	// Load pets from database if pet IDs are provided
	if len(roomCategoryDto.Pets) > 0 {
		pets, err := s.roomCategoryRepository().GetPetsByIDs(ctx, roomCategoryDto.Pets)
		if err != nil {
			s.tracer.TraceEnd(childSpan)
			return err
		}
		roomCategory.Pets = pets
	}

	// Load facilities from database if facility IDs are provided
	if len(roomCategoryDto.Facilities) > 0 {
		facilities, err := s.roomCategoryRepository().GetFacilitiesByIDs(ctx, roomCategoryDto.Facilities)
		if err != nil {
			s.tracer.TraceEnd(childSpan)
			return err
		}
		roomCategory.Facilities = facilities
	}

	// Update room category in database
	err := s.roomCategoryRepository().UpdateRoomCategory(ctx, id, roomCategory)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}

	// Delete specific images if deleteImageIds are provided
	if len(roomCategoryDto.DeleteImageIds) > 0 {
		for _, imageID := range roomCategoryDto.DeleteImageIds {
			if err := s.roomCategoryRepository().DeleteRoomCategoryImage(ctx, imageID); err != nil {
				s.logger.Error(fmt.Sprintf("Failed to delete room category image %d: %v", imageID, err))
			}
		}
	}

	// Upload new images if provided
	if len(roomCategoryDto.Images) > 0 {
		roomCategoryImages := make([]models.RoomCategoryImage, 0, len(roomCategoryDto.Images))
		basePath := "room-categories"

		for index, fileHeader := range roomCategoryDto.Images {
			timestamp := time.Now().UnixNano()
			ext := filepath.Ext(fileHeader.Filename)
			fileName := fmt.Sprintf("%d_%d_%d%s", id, timestamp, index, ext)

			if err := s.fileSystem.Put(ctx, basePath, fileName, fileHeader); err != nil {
				s.logger.Error(fmt.Sprintf("Failed to upload room category image: %v", err))
				continue
			}

			publicURL := s.fileSystem.GetPublicURL(basePath, fileName)

			roomCategoryImages = append(roomCategoryImages, models.RoomCategoryImage{
				FileName:          fileHeader.Filename,
				FilePath:          publicURL,
				FileContentType:   fileHeader.Header.Get("Content-Type"),
				FileExtensionName: strings.TrimPrefix(ext, "."),
				FileSize:          int(fileHeader.Size),
				RoomCategoryID:    &id,
			})
		}

		if len(roomCategoryImages) > 0 {
			if err = s.roomCategoryRepository().CreateRoomCategoryImages(ctx, roomCategoryImages); err != nil {
				s.logger.Error(fmt.Sprintf("Failed to save room category images: %v", err))
			}
		}
	}

	s.tracer.TraceEnd(childSpan)

	return nil
}

func (s Service) DeleteRoomCategory(ctx context.Context, id int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "DeleteRoomCategoryService", trace.WithAttributes(attribute.String("service", "DeleteRoomCategory"), attribute.Int64("id", int64(id))))

	err := s.roomCategoryRepository().DeleteRoomCategory(ctx, id)

	s.tracer.TraceEnd(childSpan)

	return err
}

func (s Service) SearchAvailableRoomCategories(ctx context.Context, propertyID int, checkInDate, checkOutDate string, numberOfPets int, petSizeId *int) ([]models.RoomCategory, map[int][]models.RoomCategoryCalendar, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "SearchAvailableRoomCategoriesService", trace.WithAttributes(attribute.String("service", "SearchAvailableRoomCategories"), attribute.Int64("propertyId", int64(propertyID))))

	// Parse dates
	checkIn, err := time.Parse("2006-01-02", checkInDate)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, nil, err
	}

	checkOut, err := time.Parse("2006-01-02", checkOutDate)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, nil, err
	}

	// Validate dates: check-out must be after check-in
	if checkOut.Before(checkIn) || checkOut.Equal(checkIn) {
		s.tracer.TraceEnd(childSpan)
		return nil, nil, exception.ErrInvalidDateRange
	}

	roomCategories, calendarsMap, err := s.roomCategoryRepository().SearchAvailableRoomCategories(ctx, propertyID, checkIn, checkOut, numberOfPets, petSizeId)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, nil, err
	}

	s.tracer.TraceEnd(childSpan)

	return roomCategories, calendarsMap, nil
}
