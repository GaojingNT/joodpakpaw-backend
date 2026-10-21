package property_module

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/utils"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
)

func (r Repository) GetPropertyPaginate(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetPropertyPaginateRepository", trace.WithAttributes(attribute.String("repository", "GetPropertyPaginate")))
		properties   = make([]models.Property, 0)
		err          error
	)

	// Get attributes
	searchAttribute, _ := pagination.GetStringAttribute("search")
	searchByAttribute, _ := pagination.GetStringAttribute("search_by")
	petTypesStr, _ := pagination.GetStringAttribute("pet_types")
	maxPriceAttr, _ := pagination.GetFloatAttribute("max_price")
	excludeHostIDStr, _ := pagination.GetStringAttribute("exclude_host_id")

	// Set tracing attributes
	r.tracer.SetAttributes(childSpan, attribute.String("search", searchAttribute))
	r.tracer.SetAttributes(childSpan, attribute.String("search_by", searchByAttribute))

	// Create pagination query
	tx := r.db.Where("is_active = ?", true)

	// Exclude properties owned by the requesting user
	if excludeHostID, err := strconv.Atoi(excludeHostIDStr); err == nil && excludeHostID > 0 {
		tx = tx.Where("host_id != ?", excludeHostID)
	}

	// maxPrice filter: properties where the cheapest room <= maxPrice
	if maxPriceAttr > 0 {
		tx = tx.Where("id IN (?)",
			r.db.Table("tbl_room_categories").
				Select("property_id").
				Where("deleted_at IS NULL").
				Group("property_id").
				Having("MIN(base_price) <= ?", maxPriceAttr))
	}

	// petTypes filter: properties that have rooms accepting ALL specified pet types (AND logic)
	if petTypesStr != "" {
		petTypes := strings.Split(petTypesStr, ",")
		for i, t := range petTypes {
			petTypes[i] = strings.ToLower(strings.TrimSpace(t))
		}
		tx = tx.Where("id IN (?)",
			r.db.Table("tbl_room_categories rc").
				Select("rc.property_id").
				Joins("JOIN tbl_room_has_pets rhp ON rhp.room_category_id = rc.id").
				Joins("JOIN tbl_pets p ON p.id = rhp.pet_id AND p.deleted_at IS NULL").
				Where("rc.deleted_at IS NULL AND LOWER(p.name) IN ?", petTypes).
				Group("rc.property_id").
				Having("COUNT(DISTINCT LOWER(p.name)) = ?", len(petTypes)))
	}

	utils.Block{
		Try: func() {
			// Execute query
			if err = tx.
				Preload("Images").
				Preload("RoomsCategory").
				Preload("RoomsCategory.Pets").
				Preload("BookingOrders.Booking").
				Scopes(models.SearchingScope(models.PropertyPublicSearchable(), searchAttribute, searchByAttribute)).
				Scopes(paginator.Paginate(properties, pagination, tx)).
				Find(&properties).Error; err != nil {
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

	// Set data
	pagination.Data = properties

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return pagination, nil
}

func (r Repository) GetAllProperties(ctx context.Context, search string, searchBy string) ([]models.Property, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetAllPropertiesRepository", trace.WithAttributes(attribute.String("repository", "GetAllProperties")))
		properties   = make([]models.Property, 0)
		err          error
	)

	// Set tracing attributes
	r.tracer.SetAttributes(childSpan, attribute.String("search", search))
	r.tracer.SetAttributes(childSpan, attribute.String("search_by", searchBy))

	// Create query
	tx := r.db

	utils.Block{
		Try: func() {
			// Execute query with preloads
			query := tx.
				Where("is_active = ?", true).
				Preload("Images").
				Preload("RoomsCategory").
				Preload("RoomsCategory.Pets").
				Preload("BookingOrders.Booking")

			// Apply search if provided
			if search != "" {
				query = query.Scopes(models.SearchingScope(models.PropertySearchable(), search, searchBy))
			}

			if err = query.Find(&properties).Error; err != nil {
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

	return properties, nil
}

func (r Repository) GetPropertyByID(ctx context.Context, id int) (models.Property, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetPropertyByIDRepository", trace.WithAttributes(attribute.String("repository", "GetPropertyByID"), attribute.Int64("id", int64(id))))
		property     models.Property
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query
			if err = r.db.
				Preload("Host").
				Preload("PropertyType").
				Preload("Images").
				Preload("RoomsCategory").
				Preload("RoomsCategory.Images").
				Preload("RoomsCategory.Pets").
				Preload("RoomsCategory.Facilities").
				Preload("Facilities").
				Preload("BookingOrders.Booking").
				Preload("BookingOrders.Booking.Guest").
				Preload("BookingOrders.Booking.ReviewImages").
				First(&property, id).Error; err != nil {
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
		return property, err
	}

	return property, nil
}

func (r Repository) CreateProperty(ctx context.Context, property *models.Property) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreatePropertyRepository", trace.WithAttributes(attribute.String("repository", "CreateProperty")))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Execute query
				if err = r.db.Create(&property).Error; err != nil {
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

func (r Repository) UpdateProperty(ctx context.Context, id int, property *models.Property) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "UpdatePropertyRepository", trace.WithAttributes(attribute.String("repository", "UpdateProperty"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Build update map dynamically to only update non-empty fields
				updateMap := make(map[string]interface{})

				if property.Title != "" {
					updateMap["title"] = property.Title
				}
				if property.Description != "" {
					updateMap["description"] = property.Description
				}
				if property.Latitude != 0 {
					updateMap["latitude"] = property.Latitude
				}
				if property.Longitude != 0 {
					updateMap["longitude"] = property.Longitude
				}
				if property.Address != "" {
					updateMap["address"] = property.Address
				}
				if property.Province != "" {
					updateMap["province"] = property.Province
				}
				if property.District != "" {
					updateMap["district"] = property.District
				}
				if property.Country != "" {
					updateMap["country"] = property.Country
				}
				if property.Zipcode != "" {
					updateMap["zipcode"] = property.Zipcode
				}
				// Note: HostID should not be updated after property creation
				if property.PropertyTypeID != nil {
					updateMap["property_type_id"] = property.PropertyTypeID
				}

				// Execute query only if there are fields to update
				if len(updateMap) > 0 {
					if err = r.db.
						Model(&models.Property{}).
						Where("id = ?", id).
						Updates(updateMap).Error; err != nil {
						return err
					}
				}

				// Update many-to-many relationships
				var existingProperty models.Property
				if err = r.db.First(&existingProperty, id).Error; err != nil {
					return err
				}

				// Update Facilities relationship
				if property.Facilities != nil {
					if err = r.db.Model(&existingProperty).Association("Facilities").Replace(property.Facilities); err != nil {
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

func (r Repository) CountRoomCategoriesByPropertyID(ctx context.Context, propertyID int) (int64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CountRoomCategoriesByPropertyIDRepository", trace.WithAttributes(attribute.String("repository", "CountRoomCategoriesByPropertyID"), attribute.Int64("propertyId", int64(propertyID))))
		count        int64
		err          error
	)

	utils.Block{
		Try: func() {
			// Count room categories for this property
			if err = r.db.
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

func (r Repository) UnbanProperty(ctx context.Context, id int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "UnbanPropertyRepository", trace.WithAttributes(attribute.String("repository", "UnbanProperty"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.Property{}).Where("id = ?", id).Updates(map[string]any{
				"is_banned_by_admin": false,
				"is_active":         false,
			}).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
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

func (r Repository) UpdatePropertyIsActive(ctx context.Context, id int, isActive bool) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "UpdatePropertyIsActiveRepository", trace.WithAttributes(attribute.String("repository", "UpdatePropertyIsActive"), attribute.Int64("id", int64(id)), attribute.Bool("isActive", isActive)))
		err          error
	)

	utils.Block{
		Try: func() {
			updates := map[string]any{"is_active": isActive}
			if isActive {
				now := time.Now()
				updates["published_at"] = now
			}
			if err = r.db.
				Model(&models.Property{}).
				Where("id = ?", id).
				Updates(updates).Error; err != nil {
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

func (r Repository) DeleteProperty(ctx context.Context, id int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "DeletePropertyRepository", trace.WithAttributes(attribute.String("repository", "DeleteProperty"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// First, check if property exists and is not deleted
				var property models.Property
				if err = tx.First(&property, id).Error; err != nil {
					return err
				}

				// Execute soft delete
				if err = tx.Delete(&property).Error; err != nil {
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

func (r Repository) GetHostAvgRating(ctx context.Context, hostID int) (float64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetHostAvgRatingRepository", trace.WithAttributes(attribute.String("repository", "GetHostAvgRating"), attribute.Int64("hostId", int64(hostID))))
		err          error
	)

	var result struct {
		AvgRating float64 `gorm:"column:avg_rating"`
	}

	if err = r.db.Model(&models.Property{}).
		Select("COALESCE(AVG(b.rating), 0)::float8 AS avg_rating").
		Joins("JOIN tbl_booking_orders bo ON bo.property_id = tbl_properties.id AND bo.deleted_at IS NULL").
		Joins("JOIN tbl_bookings b ON b.id = bo.booking_id AND b.deleted_at IS NULL AND b.rating IS NOT NULL").
		Where("tbl_properties.host_id = ?", hostID).
		Scan(&result).Error; err != nil {
		r.logger.Error("GetHostAvgRating error: " + err.Error())
		r.tracer.TraceEnd(childSpan)
		return 0, err
	}

	r.tracer.TraceEnd(childSpan)
	return result.AvgRating, nil
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

func (r Repository) CreatePropertyImages(ctx context.Context, images []models.PropertyImage) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreatePropertyImagesRepository", trace.WithAttributes(attribute.String("repository", "CreatePropertyImages")))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Execute query
				if err = r.db.Create(&images).Error; err != nil {
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

func (r Repository) GetPropertiesByHostID(ctx context.Context, hostID int) ([]models.Property, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetPropertiesByHostIDRepository", trace.WithAttributes(attribute.String("repository", "GetPropertiesByHostID"), attribute.Int64("hostId", int64(hostID))))
		properties   = make([]models.Property, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query with preloads
			if err = r.db.
				Preload("Images").
				Preload("RoomsCategory").
				Preload("RoomsCategory.Pets").
				Preload("BookingOrders.Booking").
				Where("host_id = ?", hostID).
				Find(&properties).Error; err != nil {
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

	return properties, nil
}

func (r Repository) GetPropertiesByHostIDPaginate(ctx context.Context, hostID int, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetPropertiesByHostIDPaginateRepository", trace.WithAttributes(attribute.String("repository", "GetPropertiesByHostIDPaginate"), attribute.Int64("hostId", int64(hostID))))
		properties   = make([]models.Property, 0)
		err          error
	)

	// Get attributes
	searchAttribute, _ := pagination.GetStringAttribute("search")
	searchByAttribute, _ := pagination.GetStringAttribute("search_by")

	// Set tracing attributes
	r.tracer.SetAttributes(childSpan, attribute.String("search", searchAttribute))
	r.tracer.SetAttributes(childSpan, attribute.String("search_by", searchByAttribute))
	r.tracer.SetAttributes(childSpan, attribute.Int64("hostId", int64(hostID)))

	// Create pagination query
	tx := r.db.Where("host_id = ?", hostID)

	utils.Block{
		Try: func() {
			// Execute query
			if err = tx.
				Preload("Images").
				Preload("RoomsCategory").
				Preload("RoomsCategory.Pets").
				Preload("BookingOrders.Booking").
				Scopes(models.SearchingScope(models.PropertySearchable(), searchAttribute, searchByAttribute)).
				Scopes(paginator.Paginate(properties, pagination, tx)).
				Find(&properties).Error; err != nil {
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

	// Set data
	pagination.Data = properties

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return pagination, nil
}

func (r Repository) GetPropertyForHostByID(ctx context.Context, hostID int, propertyID int) (models.Property, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetPropertyForHostByIDRepository", trace.WithAttributes(attribute.String("repository", "GetPropertyForHostByID"), attribute.Int64("hostId", int64(hostID)), attribute.Int64("propertyId", int64(propertyID))))
		property     models.Property
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query with all necessary preloads
			if err = r.db.
				Preload("Host").
				Preload("Images").
				Preload("Facilities").
				Preload("PropertyType").
				Preload("BookingOrders.Booking").
				Preload("BookingOrders.Booking.Guest").
				Preload("BookingOrders.Booking.ReviewImages").
				Preload("RoomsCategory").
				Preload("RoomsCategory.Images").
				Preload("RoomsCategory.Pets").
				Preload("RoomsCategory.Facilities").
				Where("host_id = ? AND id = ?", hostID, propertyID).
				First(&property).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Check if record not found (either property doesn't exist or doesn't belong to this host)
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
		return property, err
	}

	return property, nil
}

func (r Repository) GetPropertyImageByID(ctx context.Context, imageID int) (models.PropertyImage, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetPropertyImageByIDRepository", trace.WithAttributes(attribute.String("repository", "GetPropertyImageByID"), attribute.Int64("imageId", int64(imageID))))
		image        models.PropertyImage
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query
			if err = r.db.First(&image, imageID).Error; err != nil {
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
		return image, err
	}

	return image, nil
}

func (r Repository) DeletePropertyImage(ctx context.Context, imageID int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "DeletePropertyImageRepository", trace.WithAttributes(attribute.String("repository", "DeletePropertyImage"), attribute.Int64("imageId", int64(imageID))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// First, check if image exists and is not deleted
				var image models.PropertyImage
				if err = tx.First(&image, imageID).Error; err != nil {
					return err
				}

				// Execute soft delete
				if err = tx.Delete(&image).Error; err != nil {
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

func (r Repository) DeleteAllPropertyImages(ctx context.Context, propertyID int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "DeleteAllPropertyImagesRepository", trace.WithAttributes(attribute.String("repository", "DeleteAllPropertyImages"), attribute.Int64("propertyId", int64(propertyID))))
		err          error
	)

	utils.Block{
		Try: func() {
			// Delete all images for this property (soft delete)
			if err = r.db.Where("property_id = ?", propertyID).Delete(&models.PropertyImage{}).Error; err != nil {
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

func (r Repository) GetPropertiesByIDs(ctx context.Context, ids []int) ([]models.Property, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetPropertiesByIDsRepository", trace.WithAttributes(attribute.String("repository", "GetPropertiesByIDs")))
		properties   = make([]models.Property, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.
				Where("id IN ?", ids).
				Where("is_active = ?", true).
				Preload("Images").
				Preload("RoomsCategory").
				Preload("RoomsCategory.Pets").
				Preload("BookingOrders.Booking").
				Find(&properties).Error; err != nil {
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
		return nil, err
	}

	return properties, nil
}

func (r Repository) GetHostReviewsPaginate(ctx context.Context, hostID int, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetHostReviewsPaginateRepository", trace.WithAttributes(attribute.String("repository", "GetHostReviewsPaginate"), attribute.Int64("hostId", int64(hostID))))
		bookings     = make([]models.Booking, 0)
		err          error
	)

	// Use EXISTS to filter bookings that belong to this host's properties, avoiding duplicate rows from JOIN
	tx := r.db.
		Where(`tbl_bookings.rating IS NOT NULL AND tbl_bookings.deleted_at IS NULL AND EXISTS (
			SELECT 1 FROM tbl_booking_orders bo
			JOIN tbl_properties p ON p.id = bo.property_id AND p.deleted_at IS NULL
			WHERE bo.booking_id = tbl_bookings.id AND bo.deleted_at IS NULL AND p.host_id = ?
		)`, hostID)

	utils.Block{
		Try: func() {
			if err = tx.
				Preload("Guest").
				Preload("ReviewImages").
				Preload("BookingOrders", "deleted_at IS NULL").
				Preload("BookingOrders.Property").
				Scopes(paginator.Paginate(bookings, pagination, tx)).
				Find(&bookings).Error; err != nil {
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

	pagination.Data = bookings

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return pagination, nil
}

func (r Repository) GetFavoritedPropertyIDs(ctx context.Context, userID string) (map[int]bool, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetFavoritedPropertyIDsRepository", trace.WithAttributes(attribute.String("repository", "GetFavoritedPropertyIDs"), attribute.String("userID", userID)))
		propertyIDs  = make(map[int]bool)
		err          error
	)

	utils.Block{
		Try: func() {
			// Get all favorite folders for the user with their properties using GORM Preload
			var folders []models.FavoriteFolder
			if err = r.db.
				Where("user_id = ?", userID).
				Preload("Properties").
				Find(&folders).Error; err != nil {
				utils.Throw(err)
			}

			// Collect all property IDs from all folders
			for _, folder := range folders {
				for _, property := range folder.Properties {
					propertyIDs[property.ID] = true
				}
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

	return propertyIDs, nil
}

func cancelPropertyBookingsByIDs(tx *gorm.DB, propertyIDs []int) error {
	today := time.Now().Truncate(24 * time.Hour)

	propertyIDSet := make(map[int]struct{}, len(propertyIDs))
	for _, pid := range propertyIDs {
		propertyIDSet[pid] = struct{}{}
	}

	var bookings []models.Booking
	if err := tx.
		Preload("BookingOrders").
		Joins("JOIN tbl_booking_orders bo ON bo.booking_id = tbl_bookings.id").
		Where("bo.property_id IN ?", propertyIDs).
		Where("tbl_bookings.booking_status IN ?", []string{
			string(models.BookingStatusPending),
			string(models.BookingStatusConfirmed),
		}).
		Group("tbl_bookings.id").
		Find(&bookings).Error; err != nil {
		return err
	}

	for _, booking := range bookings {
		if booking.BookingStatus == models.BookingStatusConfirmed {
			skip := false
			for _, order := range booking.BookingOrders {
				if order.PropertyID == nil {
					continue
				}
				if _, affected := propertyIDSet[*order.PropertyID]; !affected {
					continue
				}
				if order.CheckInDate.Truncate(24 * time.Hour).Equal(today) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}

		for _, order := range booking.BookingOrders {
			for d := order.CheckInDate; d.Before(order.CheckOutDate); d = d.AddDate(0, 0, 1) {
				var calendar models.RoomCategoryCalendar
				if err := tx.Where("room_category_id = ? AND date = ?", order.RoomCategoryID, d).First(&calendar).Error; err == nil {
					calendar.AvailableRooms += order.RoomsBooked
					calendar.BookedRooms -= order.RoomsBooked
					if err := tx.Save(&calendar).Error; err != nil {
						return err
					}
				}
			}
		}
		if err := tx.Model(&booking).Update("booking_status", models.BookingStatusCancelled).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r Repository) GetActiveBookingGuestIDsByPropertyIDs(ctx context.Context, propertyIDs []int) ([]int, error) {
	var guestIDs []int
	err := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Select("DISTINCT tbl_bookings.guest_id").
		Joins("JOIN tbl_booking_orders bo ON bo.booking_id = tbl_bookings.id").
		Where("bo.property_id IN ?", propertyIDs).
		Where("tbl_bookings.booking_status IN ?", []string{
			string(models.BookingStatusPending),
			string(models.BookingStatusConfirmed),
		}).
		Pluck("guest_id", &guestIDs).Error
	return guestIDs, err
}

// resetUnbanRequestsByPropertyID soft-deletes all unban_property requests for a specific property
func resetUnbanRequestsByPropertyID(tx *gorm.DB, propertyID int) error {
	var ids []int
	if err := tx.Table("tbl_reports").
		Joins("JOIN tbl_report_categories ON tbl_report_categories.id = tbl_reports.report_category_id").
		Where("tbl_reports.property_id = ?", propertyID).
		Where("tbl_report_categories.type = ?", "unban_property").
		Where("tbl_reports.deleted_at IS NULL").
		Pluck("tbl_reports.id", &ids).Error; err != nil {
		return err
	}
	if len(ids) > 0 {
		return tx.Where("id IN ?", ids).Delete(&models.Report{}).Error
	}
	return nil
}

func (r Repository) AdminDeactivateProperty(ctx context.Context, propertyID int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "AdminDeactivatePropertyRepository", trace.WithAttributes(attribute.String("repository", "AdminDeactivateProperty"), attribute.Int64("propertyID", int64(propertyID))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				var property models.Property
				if err = tx.First(&property, propertyID).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						return exception.ErrRecordNotFound
					}
					return err
				}
				if err = tx.Model(&models.Property{}).Where("id = ?", propertyID).Updates(map[string]any{
					"is_active":          false,
					"is_banned_by_admin": true,
				}).Error; err != nil {
					return err
				}
				if err = cancelPropertyBookingsByIDs(tx, []int{propertyID}); err != nil {
					return err
				}
				return resetUnbanRequestsByPropertyID(tx, propertyID)
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == exception.ErrRecordNotFound {
				// pass through
			} else {
				err = e.(error)
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)
	return err
}

func (r Repository) CountUnbanRequestsByPropertyIDs(ctx context.Context, propertyIDs []int) (map[int]int64, error) {
	type result struct {
		PropertyID int   `gorm:"column:property_id"`
		Count      int64 `gorm:"column:count"`
	}
	var rows []result
	err := r.db.WithContext(ctx).
		Table("tbl_reports").
		Select("tbl_reports.property_id, COUNT(*) as count").
		Joins("JOIN tbl_report_categories ON tbl_report_categories.id = tbl_reports.report_category_id").
		Where("tbl_report_categories.type = ?", "unban_property").
		Where("tbl_reports.property_id IN ?", propertyIDs).
		Where("tbl_reports.deleted_at IS NULL").
		Group("tbl_reports.property_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	counts := make(map[int]int64, len(rows))
	for _, row := range rows {
		counts[row.PropertyID] = row.Count
	}
	return counts, nil
}

func (r Repository) GetPendingUnbanPropertyIDs(ctx context.Context, propertyIDs []int) (map[int]bool, error) {
	var ids []int
	err := r.db.WithContext(ctx).
		Table("tbl_reports").
		Select("DISTINCT tbl_reports.property_id").
		Joins("JOIN tbl_report_categories ON tbl_report_categories.id = tbl_reports.report_category_id").
		Where("tbl_report_categories.type = ?", "unban_property").
		Where("tbl_reports.property_id IN ?", propertyIDs).
		Where("tbl_reports.status = ?", "pending").
		Where("tbl_reports.deleted_at IS NULL").
		Pluck("tbl_reports.property_id", &ids).Error
	if err != nil {
		return nil, err
	}
	pending := make(map[int]bool, len(ids))
	for _, id := range ids {
		pending[id] = true
	}
	return pending, nil
}

func (r Repository) CountUnbanRequestsByPropertyID(ctx context.Context, propertyID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("tbl_reports").
		Joins("JOIN tbl_report_categories ON tbl_report_categories.id = tbl_reports.report_category_id").
		Where("tbl_report_categories.type = ?", "unban_property").
		Where("tbl_reports.property_id = ?", propertyID).
		Where("tbl_reports.deleted_at IS NULL").
		Count(&count).Error
	return count, err
}

func (r Repository) HasPendingUnbanRequestByPropertyID(ctx context.Context, propertyID int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("tbl_reports").
		Joins("JOIN tbl_report_categories ON tbl_report_categories.id = tbl_reports.report_category_id").
		Where("tbl_report_categories.type = ?", "unban_property").
		Where("tbl_reports.property_id = ?", propertyID).
		Where("tbl_reports.status = ?", "pending").
		Where("tbl_reports.deleted_at IS NULL").
		Count(&count).Error
	return count > 0, err
}

func (r Repository) CountUnbanAllPropertyRequests(ctx context.Context, hostID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("tbl_reports").
		Joins("JOIN tbl_report_categories ON tbl_report_categories.id = tbl_reports.report_category_id").
		Where("tbl_reports.reporter_id = ?", hostID).
		Where("tbl_report_categories.type = ?", "unban_property").
		Where("tbl_reports.property_id IS NULL").
		Where("tbl_reports.deleted_at IS NULL").
		Count(&count).Error
	return count, err
}

func (r Repository) HasPendingUnbanAllPropertyRequest(ctx context.Context, hostID int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("tbl_reports").
		Joins("JOIN tbl_report_categories ON tbl_report_categories.id = tbl_reports.report_category_id").
		Where("tbl_reports.reporter_id = ?", hostID).
		Where("tbl_report_categories.type = ?", "unban_property").
		Where("tbl_reports.property_id IS NULL").
		Where("tbl_reports.status = ?", "pending").
		Where("tbl_reports.deleted_at IS NULL").
		Count(&count).Error
	return count > 0, err
}

func (r Repository) AdminGetPropertiesPaginate(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "AdminGetPropertiesPaginateRepository", trace.WithAttributes(attribute.String("repository", "AdminGetPropertiesPaginate")))
		properties   = make([]models.Property, 0)
		err          error
	)

	searchAttribute, _ := pagination.GetStringAttribute("search")
	searchByAttribute, _ := pagination.GetStringAttribute("search_by")

	tx := r.db

	utils.Block{
		Try: func() {
			if err = tx.
				Preload("Host").
				Scopes(models.SearchingScope(models.PropertySearchable(), searchAttribute, searchByAttribute)).
				Scopes(paginator.Paginate(properties, pagination, tx)).
				Find(&properties).Error; err != nil {
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

	pagination.Data = properties
	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}
	return pagination, nil
}

func (r Repository) AdminGetPropertyByID(ctx context.Context, id int) (models.Property, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "AdminGetPropertyByIDRepository", trace.WithAttributes(attribute.String("repository", "AdminGetPropertyByID"), attribute.Int64("id", int64(id))))
		property     models.Property
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.
				Preload("Host").
				First(&property, id).Error; err != nil {
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
		return property, err
	}
	return property, nil
}

// GetTopRatedPropertyIDs returns active property IDs ordered by avg rating DESC,
// filtered to only include properties with avgRating >= minRating.
func (r Repository) GetTopRatedPropertyIDs(ctx context.Context, minRating float64) ([]int, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetTopRatedPropertyIDsRepository", trace.WithAttributes(attribute.String("repository", "GetTopRatedPropertyIDs")))
		ids          []int
		err          error
	)

	err = r.db.WithContext(ctx).
		Table("tbl_properties p").
		Select("p.id").
		Joins("JOIN tbl_booking_orders bo ON bo.property_id = p.id AND bo.deleted_at IS NULL").
		Joins("JOIN tbl_bookings b ON b.id = bo.booking_id AND b.deleted_at IS NULL AND b.rating IS NOT NULL").
		Where("p.is_active = ? AND p.deleted_at IS NULL", true).
		Group("p.id").
		Having("COALESCE(AVG(b.rating), 0) >= ?", minRating).
		Order("AVG(b.rating) DESC").
		Pluck("p.id", &ids).Error

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		r.logger.Error("GetTopRatedPropertyIDs error: " + err.Error())
		sentry.CaptureException(err)
		return nil, exception.ErrDbQueryStatement
	}
	return ids, nil
}

// GetBestValuePropertyIDs returns active property IDs where the property's minimum
// room base_price is at or below the global average of minimum prices across all properties.
// Results are ordered by min base_price ASC.
func (r Repository) GetBestValuePropertyIDs(ctx context.Context) ([]int, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetBestValuePropertyIDsRepository", trace.WithAttributes(attribute.String("repository", "GetBestValuePropertyIDs")))
		ids          []int
		err          error
	)

	err = r.db.WithContext(ctx).
		Table("tbl_properties p").
		Select("p.id").
		Joins("JOIN (SELECT property_id, MIN(base_price) as min_price FROM tbl_room_categories WHERE deleted_at IS NULL GROUP BY property_id) rc ON rc.property_id = p.id").
		Where("p.is_active = ? AND p.deleted_at IS NULL", true).
		Where("rc.min_price <= (SELECT AVG(sub.min_price) FROM (SELECT MIN(base_price) as min_price FROM tbl_room_categories WHERE deleted_at IS NULL GROUP BY property_id) sub)").
		Order("rc.min_price ASC").
		Pluck("p.id", &ids).Error

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		r.logger.Error("GetBestValuePropertyIDs error: " + err.Error())
		sentry.CaptureException(err)
		return nil, exception.ErrDbQueryStatement
	}
	return ids, nil
}

// GetNewListingsPropertyIDs returns active property IDs published within the last N days,
// ordered by published_at DESC.
func (r Repository) GetNewListingsPropertyIDs(ctx context.Context, days int) ([]int, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetNewListingsPropertyIDsRepository", trace.WithAttributes(attribute.String("repository", "GetNewListingsPropertyIDs")))
		ids          []int
		err          error
	)

	cutoff := time.Now().AddDate(0, 0, -days)

	err = r.db.WithContext(ctx).
		Model(&models.Property{}).
		Where("is_active = ? AND deleted_at IS NULL AND published_at >= ?", true, cutoff).
		Order("published_at DESC").
		Pluck("id", &ids).Error

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		r.logger.Error("GetNewListingsPropertyIDs error: " + err.Error())
		sentry.CaptureException(err)
		return nil, exception.ErrDbQueryStatement
	}
	return ids, nil
}

// GetRecentActivePropertyIDs returns the N most recently published active property IDs,
// ordered by published_at DESC. Used as fallback for new listings.
func (r Repository) GetRecentActivePropertyIDs(ctx context.Context, limit int) ([]int, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetRecentActivePropertyIDsRepository", trace.WithAttributes(attribute.String("repository", "GetRecentActivePropertyIDs")))
		ids          []int
		err          error
	)

	err = r.db.WithContext(ctx).
		Model(&models.Property{}).
		Where("is_active = ? AND deleted_at IS NULL AND published_at IS NOT NULL", true).
		Order("published_at DESC").
		Limit(limit).
		Pluck("id", &ids).Error

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		r.logger.Error("GetRecentActivePropertyIDs error: " + err.Error())
		sentry.CaptureException(err)
		return nil, exception.ErrDbQueryStatement
	}
	return ids, nil
}

// GetTrendingPropertyIDs returns active property IDs that have at least one booking
// in the last 30 days, ordered by that booking count DESC.
func (r Repository) GetTrendingPropertyIDs(ctx context.Context) ([]int, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetTrendingPropertyIDsRepository", trace.WithAttributes(attribute.String("repository", "GetTrendingPropertyIDs")))
		ids          []int
		err          error
	)

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	err = r.db.WithContext(ctx).
		Table("tbl_properties p").
		Select("p.id").
		Joins("JOIN tbl_booking_orders bo ON bo.property_id = p.id AND bo.deleted_at IS NULL").
		Joins("JOIN tbl_bookings b ON b.id = bo.booking_id AND b.deleted_at IS NULL").
		Where("p.is_active = ? AND p.deleted_at IS NULL", true).
		Where("b.created_at >= ?", thirtyDaysAgo).
		Group("p.id").
		Order("COUNT(DISTINCT b.id) DESC").
		Pluck("p.id", &ids).Error

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		r.logger.Error("GetTrendingPropertyIDs error: " + err.Error())
		sentry.CaptureException(err)
		return nil, exception.ErrDbQueryStatement
	}
	return ids, nil
}

// GetPropertyIDsByHostID returns all property IDs owned by the given host.
func (r Repository) GetPropertyIDsByHostID(ctx context.Context, hostID int) ([]int, error) {
	var ids []int
	err := r.db.WithContext(ctx).
		Model(&models.Property{}).
		Where("host_id = ? AND deleted_at IS NULL", hostID).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// GetPropertyIDsForFilter returns a set of property IDs matching the given petTypes and/or
// maxPrice filters. Returns nil when no filters are active (meaning: no restriction).
func (r Repository) GetPropertyIDsForFilter(ctx context.Context, petTypes []string, maxPrice float64) (map[int]bool, error) {
	_, childSpan := r.tracer.TraceStart(ctx, "GetPropertyIDsForFilterRepository", trace.WithAttributes(attribute.String("repository", "GetPropertyIDsForFilter")))
	defer r.tracer.TraceEnd(childSpan)

	if len(petTypes) == 0 && maxPrice <= 0 {
		return nil, nil
	}

	var petSet, priceSet map[int]bool

	if len(petTypes) > 0 {
		lowered := make([]string, len(petTypes))
		for i, t := range petTypes {
			lowered[i] = strings.ToLower(strings.TrimSpace(t))
		}
		var ids []int
		err := r.db.WithContext(ctx).
			Table("tbl_room_categories rc").
			Select("rc.property_id").
			Joins("JOIN tbl_room_has_pets rhp ON rhp.room_category_id = rc.id").
			Joins("JOIN tbl_pets p ON p.id = rhp.pet_id AND p.deleted_at IS NULL").
			Where("rc.deleted_at IS NULL AND LOWER(p.name) IN ?", lowered).
			Group("rc.property_id").
			Having("COUNT(DISTINCT LOWER(p.name)) = ?", len(lowered)).
			Pluck("rc.property_id", &ids).Error
		if err != nil {
			r.logger.Error("GetPropertyIDsForFilter (petTypes) error: " + err.Error())
			sentry.CaptureException(err)
			return nil, exception.ErrDbQueryStatement
		}
		petSet = make(map[int]bool, len(ids))
		for _, id := range ids {
			petSet[id] = true
		}
	}

	if maxPrice > 0 {
		var ids []int
		err := r.db.WithContext(ctx).
			Table("tbl_room_categories").
			Select("property_id").
			Where("deleted_at IS NULL").
			Group("property_id").
			Having("MIN(base_price) <= ?", maxPrice).
			Pluck("property_id", &ids).Error
		if err != nil {
			r.logger.Error("GetPropertyIDsForFilter (maxPrice) error: " + err.Error())
			sentry.CaptureException(err)
			return nil, exception.ErrDbQueryStatement
		}
		priceSet = make(map[int]bool, len(ids))
		for _, id := range ids {
			priceSet[id] = true
		}
	}

	// Both filters: return intersection
	if petSet != nil && priceSet != nil {
		result := make(map[int]bool)
		for id := range petSet {
			if priceSet[id] {
				result[id] = true
			}
		}
		return result, nil
	}
	if petSet != nil {
		return petSet, nil
	}
	return priceSet, nil
}
