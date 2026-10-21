package favorite_folder_module

import (
	"context"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/utils"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
)

// GetAllFavoriteFolders retrieves all favorite folders without pagination
func (r Repository) GetAllFavoriteFolders(ctx context.Context, userID string) ([]models.FavoriteFolder, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetAllFavoriteFoldersRepository", trace.WithAttributes(attribute.String("repository", "GetAllFavoriteFolders")))
		folders      = make([]models.FavoriteFolder, 0)
		err          error
	)

	// Set tracing attributes
	r.tracer.SetAttributes(childSpan, attribute.String("userID", userID))

	utils.Block{
		Try: func() {
			// Execute query - Preload only active Properties with their first Image
			if err = r.db.
				Where("user_id = ?", userID).
				Preload("Properties", func(db *gorm.DB) *gorm.DB {
					return db.Where("is_active = ?", true).Order("id ASC")
				}).
				Preload("Properties.Images", func(db *gorm.DB) *gorm.DB {
					return db.Order("id ASC")
				}).
				Find(&folders).Error; err != nil {
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

	return folders, nil
}

// GetFavoriteFoldersPaginate retrieves favorite folders with pagination
func (r Repository) GetFavoriteFoldersPaginate(ctx context.Context, pagination *paginator.Pagination, userID string) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetFavoriteFoldersPaginateRepository", trace.WithAttributes(attribute.String("repository", "GetFavoriteFoldersPaginate")))
		folders      = make([]models.FavoriteFolder, 0)
		err          error
	)

	// Set tracing attributes
	r.tracer.SetAttributes(childSpan, attribute.String("userID", userID))

	// Create pagination query
	tx := r.db.Where("user_id = ?", userID)

	utils.Block{
		Try: func() {
			// Execute query - Preload only active Properties with their Images
			if err = tx.
				Preload("Properties", func(db *gorm.DB) *gorm.DB {
					return db.Where("is_active = ?", true).Order("id ASC")
				}).
				Preload("Properties.Images", func(db *gorm.DB) *gorm.DB {
					return db.Order("id ASC")
				}).
				Scopes(paginator.Paginate(folders, pagination, tx)).
				Find(&folders).Error; err != nil {
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
	pagination.Data = folders

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return pagination, nil
}

// GetFavoriteFolderById retrieves a favorite folder by ID with all properties
func (r Repository) GetFavoriteFolderById(ctx context.Context, id int, userID string) (*models.FavoriteFolder, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetFavoriteFolderByIdRepository", trace.WithAttributes(attribute.String("repository", "GetFavoriteFolderById")))
		folder       models.FavoriteFolder
		err          error
	)

	// Set tracing attributes
	r.tracer.SetAttributes(childSpan, attribute.Int("id", id))
	r.tracer.SetAttributes(childSpan, attribute.String("userID", userID))

	utils.Block{
		Try: func() {
			// Execute query - Preload only active Properties with their Images (ordered by ID)
			if err = r.db.
				Where("id = ? AND user_id = ?", id, userID).
				Preload("Properties", func(db *gorm.DB) *gorm.DB {
					return db.Where("is_active = ?", true).Order("id ASC")
				}).
				Preload("Properties.Images", func(db *gorm.DB) *gorm.DB {
					return db.Order("id ASC")
				}).
				First(&folder).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)

			// Handle specific errors
			if err.Error() == "record not found" {
				err = exception.ErrRecordNotFound
			} else {
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

	return &folder, nil
}

// CreateFavoriteFolder creates a new favorite folder
func (r Repository) CreateFavoriteFolder(ctx context.Context, name string, userID string) (*models.FavoriteFolder, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateFavoriteFolderRepository", trace.WithAttributes(attribute.String("repository", "CreateFavoriteFolder")))
		folder       models.FavoriteFolder
		err          error
	)

	// Set tracing attributes
	r.tracer.SetAttributes(childSpan, attribute.String("name", name))
	r.tracer.SetAttributes(childSpan, attribute.String("userID", userID))

	folder.Name = name
	folder.UserID = &userID

	utils.Block{
		Try: func() {
			if err = r.db.Create(&folder).Error; err != nil {
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

	return &folder, nil
}

// GetFavoriteFolderCountByUserId counts the number of favorite folders for a user
func (r Repository) GetFavoriteFolderCountByUserId(ctx context.Context, userID string) (int64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetFavoriteFolderCountByUserIdRepository", trace.WithAttributes(attribute.String("repository", "GetFavoriteFolderCountByUserId")))
		count        int64
		err          error
	)

	r.tracer.SetAttributes(childSpan, attribute.String("userID", userID))

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.FavoriteFolder{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
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
		return 0, err
	}

	return count, nil
}

// ValidateActiveProperty checks if a property exists and is active
func (r Repository) ValidateActiveProperty(ctx context.Context, propertyID int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "ValidateActivePropertyRepository", trace.WithAttributes(attribute.String("repository", "ValidateActiveProperty")))
		err          error
	)

	r.tracer.SetAttributes(childSpan, attribute.Int("propertyID", propertyID))

	utils.Block{
		Try: func() {
			var property models.Property
			if err = r.db.Where("id = ? AND is_active = ?", propertyID, true).First(&property).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)

			if err.Error() == "record not found" {
				err = exception.ErrRecordNotFound
			} else {
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	return err
}

// AddPropertyToFolder adds a property to a favorite folder
func (r Repository) AddPropertyToFolder(ctx context.Context, folderID int, propertyID int, userID string) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "AddPropertyToFolderRepository", trace.WithAttributes(attribute.String("repository", "AddPropertyToFolder")))
		err          error
	)

	r.tracer.SetAttributes(childSpan, attribute.Int("folderID", folderID))
	r.tracer.SetAttributes(childSpan, attribute.Int("propertyID", propertyID))
	r.tracer.SetAttributes(childSpan, attribute.String("userID", userID))

	utils.Block{
		Try: func() {
			// Verify folder belongs to user and property exists
			var folder models.FavoriteFolder
			if err = r.db.Where("id = ? AND user_id = ?", folderID, userID).First(&folder).Error; err != nil {
				utils.Throw(err)
			}

			var property models.Property
			if err = r.db.Where("id = ? AND is_active = ?", propertyID, true).First(&property).Error; err != nil {
				utils.Throw(err)
			}

			// Check if property already exists in any folder of this user
			var existingFolders []models.FavoriteFolder
			if err = r.db.
				Where("user_id = ?", userID).
				Preload("Properties", "id = ?", propertyID).
				Find(&existingFolders).Error; err != nil {
				utils.Throw(err)
			}

			// Check if any folder contains this property
			for _, existingFolder := range existingFolders {
				if len(existingFolder.Properties) > 0 {
					// Check if it's the same folder
					if existingFolder.ID == folderID {
						// Property already exists in the same folder
						utils.Throw(exception.ErrPropertyAlreadyInSameFolder)
					} else {
						// Property already exists in a different folder
						utils.Throw(exception.ErrPropertyAlreadyInFolder)
					}
				}
			}

			// Add property to folder using Association
			if err = r.db.Model(&folder).Association("Properties").Append(&property); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)

			if err.Error() == "record not found" {
				err = exception.ErrRecordNotFound
			} else if err == exception.ErrPropertyAlreadyInFolder {
				// Keep the error as is, don't convert to DbQueryStatement
				err = exception.ErrPropertyAlreadyInFolder
			} else if err == exception.ErrPropertyAlreadyInSameFolder {
				// Keep the error as is, don't convert to DbQueryStatement
				err = exception.ErrPropertyAlreadyInSameFolder
			} else {
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	return err
}

// GetFavoriteFolderWithProperties retrieves a folder with properties count
func (r Repository) GetFavoriteFolderWithProperties(ctx context.Context, folderID int, userID string) (*models.FavoriteFolder, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetFavoriteFolderWithPropertiesRepository", trace.WithAttributes(attribute.String("repository", "GetFavoriteFolderWithProperties")))
		folder       models.FavoriteFolder
		err          error
	)

	r.tracer.SetAttributes(childSpan, attribute.Int("folderID", folderID))
	r.tracer.SetAttributes(childSpan, attribute.String("userID", userID))

	utils.Block{
		Try: func() {
			if err = r.db.
				Where("id = ? AND user_id = ?", folderID, userID).
				Preload("Properties", "is_active = ?", true).
				First(&folder).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)

			if err.Error() == "record not found" {
				err = exception.ErrRecordNotFound
			} else {
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return &folder, nil
}

// UpdateFolderName updates the name of a favorite folder
func (r Repository) UpdateFolderName(ctx context.Context, folderID int, name string, userID string) (*models.FavoriteFolder, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "UpdateFolderNameRepository", trace.WithAttributes(attribute.String("repository", "UpdateFolderName")))
		folder       models.FavoriteFolder
		err          error
	)

	r.tracer.SetAttributes(childSpan, attribute.Int("folderID", folderID))
	r.tracer.SetAttributes(childSpan, attribute.String("name", name))
	r.tracer.SetAttributes(childSpan, attribute.String("userID", userID))

	utils.Block{
		Try: func() {
			// Verify folder belongs to user
			if err = r.db.Where("id = ? AND user_id = ?", folderID, userID).First(&folder).Error; err != nil {
				utils.Throw(err)
			}

			// Update folder name
			folder.Name = name
			if err = r.db.Save(&folder).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)

			if err.Error() == "record not found" {
				err = exception.ErrRecordNotFound
			} else {
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return &folder, nil
}

// DeleteFolder soft deletes a favorite folder
func (r Repository) DeleteFolder(ctx context.Context, folderID int, userID string) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "DeleteFolderRepository", trace.WithAttributes(attribute.String("repository", "DeleteFolder")))
		err          error
	)

	r.tracer.SetAttributes(childSpan, attribute.Int("folderID", folderID))
	r.tracer.SetAttributes(childSpan, attribute.String("userID", userID))

	utils.Block{
		Try: func() {
			// Verify folder belongs to user
			var folder models.FavoriteFolder
			if err = r.db.Where("id = ? AND user_id = ?", folderID, userID).First(&folder).Error; err != nil {
				utils.Throw(err)
			}

			// Soft delete folder
			if err = r.db.Delete(&folder).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)

			if err.Error() == "record not found" {
				err = exception.ErrRecordNotFound
			} else {
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	return err
}

// RemovePropertyFromAllUserFolders removes a property from all favorite folders of a user
// If a folder becomes empty after removal, it will be soft deleted
func (r Repository) RemovePropertyFromAllUserFolders(ctx context.Context, propertyID int, userID string) (int, int, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "RemovePropertyFromAllUserFoldersRepository", trace.WithAttributes(attribute.String("repository", "RemovePropertyFromAllUserFolders")))
		err          error
		foldersAffected int
		foldersDeleted int
	)

	r.tracer.SetAttributes(childSpan, attribute.Int("propertyID", propertyID))
	r.tracer.SetAttributes(childSpan, attribute.String("userID", userID))

	utils.Block{
		Try: func() {
			// Verify property exists
			var property models.Property
			if err = r.db.First(&property, propertyID).Error; err != nil {
				utils.Throw(err)
			}

			// Get all folders of the user
			var folders []models.FavoriteFolder
			if err = r.db.
				Where("user_id = ?", userID).
				Preload("Properties", "id = ?", propertyID).
				Find(&folders).Error; err != nil {
				utils.Throw(err)
			}

			// Remove property from each folder using GORM Association API
			for _, folder := range folders {
				// Check if this folder actually contains the property
				if len(folder.Properties) > 0 {
					if err = r.db.Model(&folder).Association("Properties").Delete(&property); err != nil {
						utils.Throw(err)
					}
					foldersAffected++

					// Check if folder is now empty after removal
					remainingCount := r.db.Model(&folder).Association("Properties").Count()

					// If folder is empty, soft delete it
					if remainingCount == 0 {
						if err = r.db.Delete(&folder).Error; err != nil {
							utils.Throw(err)
						}
						foldersDeleted++
					}
				}
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)

			if err.Error() == "record not found" {
				err = exception.ErrRecordNotFound
			} else {
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return 0, 0, err
	}

	return foldersAffected, foldersDeleted, nil
}

