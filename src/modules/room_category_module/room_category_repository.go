package room_category_module

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/utils"
	"joodpakpaw/src/models"
)

func (r Repository) GetAllRoomCategories(ctx context.Context) ([]models.RoomCategory, error) {
	var (
		_, childSpan   = r.tracer.TraceStart(ctx, "GetAllRoomCategoriesRepository", trace.WithAttributes(attribute.String("repository", "GetAllRoomCategories")))
		roomCategories []models.RoomCategory
		err            error
	)

	utils.Block{
		Try: func() {
			// Get all room categories with all relationships preloaded
			if err = r.db.
				Preload("Images").
				Preload("Pets").
				Preload("Facilities").
				Preload("PetSizeCategory").
				Find(&roomCategories).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return roomCategories, nil
}

func (r Repository) GetRoomCategoriesByPropertyID(ctx context.Context, propertyID int) ([]models.RoomCategory, error) {
	var (
		_, childSpan   = r.tracer.TraceStart(ctx, "GetRoomCategoriesByPropertyIDRepository", trace.WithAttributes(attribute.String("repository", "GetRoomCategoriesByPropertyID"), attribute.Int64("propertyId", int64(propertyID))))
		roomCategories = make([]models.RoomCategory, 0)
		err            error
	)

	utils.Block{
		Try: func() {
			// Get room categories by property ID with all relationships preloaded
			if err = r.db.
				Where("property_id = ?", propertyID).
				Preload("Pets").
				Preload("Facilities").
				Preload("PetSizeCategory").
				Preload("Images").
				Find(&roomCategories).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return roomCategories, nil
}

func (r Repository) GetRoomCategoryByID(ctx context.Context, id int) (models.RoomCategory, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetRoomCategoryByIDRepository", trace.WithAttributes(attribute.String("repository", "GetRoomCategoryByID"), attribute.Int64("id", int64(id))))
		roomCategory models.RoomCategory
		err          error
	)

	utils.Block{
		Try: func() {
			// Get room category by ID with all relationships preloaded
			if err = r.db.
				Preload("Images").
				Preload("Pets").
				Preload("Facilities").
				Preload("PetSizeCategory").
				First(&roomCategory, id).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return roomCategory, err
	}

	return roomCategory, nil
}

func (r Repository) CreateRoomCategory(ctx context.Context, roomCategory *models.RoomCategory) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateRoomCategoryRepository", trace.WithAttributes(attribute.String("repository", "CreateRoomCategory")))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Create room category
				if err = r.db.Create(&roomCategory).Error; err != nil {
					return err
				}

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return err
	}

	return nil
}

func (r Repository) DeleteRoomCategory(ctx context.Context, id int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "DeleteRoomCategoryRepository", trace.WithAttributes(attribute.String("repository", "DeleteRoomCategory"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// First, check if room category exists and is not deleted
				var roomCategory models.RoomCategory
				if err = tx.First(&roomCategory, id).Error; err != nil {
					return err
				}

				// Execute soft delete
				if err = tx.Delete(&roomCategory).Error; err != nil {
					return err
				}

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return err
	}

	return nil
}

func (r Repository) GetPetsByIDs(ctx context.Context, petIDs []int) ([]models.Pet, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetPetsByIDsRepository", trace.WithAttributes(attribute.String("repository", "GetPetsByIDs")))
		pets         = make([]models.Pet, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query to find pets by IDs
			if err = r.db.Where("id IN ?", petIDs).Find(&pets).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return pets, nil
}

func (r Repository) GetFacilitiesByIDs(ctx context.Context, facilityIDs []int) ([]models.Facility, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetFacilitiesByIDsRepository", trace.WithAttributes(attribute.String("repository", "GetFacilitiesByIDs")))
		facilities   = make([]models.Facility, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query to find facilities by IDs
			if err = r.db.Where("id IN ?", facilityIDs).Find(&facilities).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return facilities, nil
}

func (r Repository) UpdateRoomCategory(ctx context.Context, id int, roomCategory *models.RoomCategory) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "UpdateRoomCategoryRepository", trace.WithAttributes(attribute.String("repository", "UpdateRoomCategory"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Build update map dynamically to only update non-empty fields
				updateMap := make(map[string]interface{})

				if roomCategory.Name != "" {
					updateMap["name"] = roomCategory.Name
				}
				if roomCategory.TotalRoom > 0 {
					updateMap["total_room"] = roomCategory.TotalRoom
				}
				if roomCategory.NumberOfPets >= 0 {
					updateMap["number_of_pets"] = roomCategory.NumberOfPets
				}
				if roomCategory.BasePrice >= 0 {
					updateMap["base_price"] = roomCategory.BasePrice
				}
				if roomCategory.PetSizeCategoryID != nil {
					updateMap["pet_size_category_id"] = roomCategory.PetSizeCategoryID
				}

				// Execute query only if there are fields to update
				if len(updateMap) > 0 {
					if err = r.db.
						Model(&models.RoomCategory{}).
						Where("id = ?", id).
						Updates(updateMap).Error; err != nil {
						return err
					}
				}

				// Update many-to-many relationships
				var existingRoomCategory models.RoomCategory
				if err = r.db.First(&existingRoomCategory, id).Error; err != nil {
					return err
				}

				// Update Pets relationship
				if roomCategory.Pets != nil {
					if err = r.db.Model(&existingRoomCategory).Association("Pets").Replace(roomCategory.Pets); err != nil {
						return err
					}
				}

				// Update Facilities relationship
				if roomCategory.Facilities != nil {
					if err = r.db.Model(&existingRoomCategory).Association("Facilities").Replace(roomCategory.Facilities); err != nil {
						return err
					}
				}

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return err
	}

	return nil
}

func (r Repository) CreateRoomCategoryWithCalendar(ctx context.Context, roomCategory *models.RoomCategory) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateRoomCategoryWithCalendarRepository", trace.WithAttributes(attribute.String("repository", "CreateRoomCategoryWithCalendar")))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Create room category
				if err = tx.Create(&roomCategory).Error; err != nil {
					return err
				}

				// Generate 365 days of calendar entries
				calendars := make([]models.RoomCategoryCalendar, 365)
				startDate := time.Now()

				for i := 0; i < 365; i++ {
					date := startDate.AddDate(0, 0, i)
					calendars[i] = models.RoomCategoryCalendar{
						Date:             date,
						AvailableRooms:   roomCategory.TotalRoom,
						BookedRooms:      0,
						UnavailableRooms: 0,
						PriceAdjustment:  0,
						RoomCategoryID:   &roomCategory.ID,
					}
				}

				// Batch insert calendar entries
				if err = tx.CreateInBatches(calendars, 100).Error; err != nil {
					return err
				}

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return err
	}

	return nil
}

func (r Repository) SearchAvailableRoomCategories(ctx context.Context, propertyID int, checkInDate, checkOutDate time.Time, numberOfPets int, petSizeId *int) ([]models.RoomCategory, map[int][]models.RoomCategoryCalendar, error) {
	var (
		_, childSpan   = r.tracer.TraceStart(ctx, "SearchAvailableRoomCategoriesRepository", trace.WithAttributes(attribute.String("repository", "SearchAvailableRoomCategories"), attribute.Int64("propertyId", int64(propertyID))))
		roomCategories = make([]models.RoomCategory, 0)
		calendarsMap   = make(map[int][]models.RoomCategoryCalendar)
		err            error
	)

	utils.Block{
		Try: func() {
			// Validate: check-out date must be after check-in date
			if !checkOutDate.After(checkInDate) {
				err = exception.ErrInvalidDateRange
				utils.Throw(err)
			}

			// Build query for room categories
			query := r.db.
				Where("property_id = ?", propertyID).
				Where("number_of_pets > ?", 0)

			// Filter by pet size: show rooms whose pet_size_category code >= the code of the given pet size id
			// small id  → code >= 0 → all rooms, medium id → code >= 1 → medium+large, large id → code >= 2 → large only
			if petSizeId != nil {
				query = query.Where("pet_size_category_id IN (SELECT id FROM tbl_pet_size_categories WHERE code >= (SELECT code FROM tbl_pet_size_categories WHERE id = ? AND deleted_at IS NULL) AND deleted_at IS NULL)", *petSizeId)
			}

			// Get room categories with relationships
			if err = query.
				Preload("Pets").
				Preload("Facilities").
				Preload("PetSizeCategory").
				Preload("Property").
				Find(&roomCategories).Error; err != nil {
				utils.Throw(err)
			}

			// For each room category, get calendars (return all categories, not just available)
			for _, roomCategory := range roomCategories {
				var calendars []models.RoomCategoryCalendar

				// Get calendars for date range
				if err = r.db.
					Where("room_category_id = ?", roomCategory.ID).
					Where("date >= ?", checkInDate).
					Where("date < ?", checkOutDate).
					Order("date ASC").
					Find(&calendars).Error; err != nil {
					utils.Throw(err)
				}

				// Store calendars for this room category
				calendarsMap[roomCategory.ID] = calendars
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, nil, err
	}

	return roomCategories, calendarsMap, nil
}

// CountRoomCategoriesByPropertyID counts the number of room categories for a specific property
func (r Repository) CountRoomCategoriesByPropertyID(ctx context.Context, propertyID int) (int64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CountRoomCategoriesByPropertyIDRepository", trace.WithAttributes(attribute.String("repository", "CountRoomCategoriesByPropertyID"), attribute.Int("propertyId", propertyID)))
		count        int64
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.WithContext(ctx).
				Model(&models.RoomCategory{}).
				Where("property_id = ?", propertyID).
				Count(&count).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r Repository) DeleteRoomCategoryImage(ctx context.Context, imageID int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "DeleteRoomCategoryImageRepository", trace.WithAttributes(attribute.String("repository", "DeleteRoomCategoryImage"), attribute.Int64("imageId", int64(imageID))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				var image models.RoomCategoryImage
				if err = tx.First(&image, imageID).Error; err != nil {
					return err
				}
				return tx.Delete(&image).Error
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return err
	}
	return nil
}

func (r Repository) CreateRoomCategoryImages(ctx context.Context, images []models.RoomCategoryImage) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateRoomCategoryImagesRepository", trace.WithAttributes(attribute.String("repository", "CreateRoomCategoryImages")))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Create(&images).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return err
	}
	return nil
}
