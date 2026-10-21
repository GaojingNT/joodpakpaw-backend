package favorite_folder_module

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/favorite_folder_module/dtos"
	"joodpakpaw/src/modules/favorite_folder_module/responses"
)

// GetAllFavoriteFolders retrieves all favorite folders without pagination
func (s Service) GetAllFavoriteFolders(ctx context.Context, userID string) (*responses.GetFavoriteFoldersResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetAllFavoriteFoldersService", trace.WithAttributes(attribute.String("service", "GetAllFavoriteFolders")))

	folders, err := s.favoriteFolderRepository().GetAllFavoriteFolders(ctx, userID)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	// Convert to response
	response := new(responses.GetFavoriteFoldersResponse).Make(folders)

	s.tracer.TraceEnd(childSpan)

	return response, nil
}

// GetFavoriteFoldersPaginate retrieves favorite folders with pagination
func (s Service) GetFavoriteFoldersPaginate(ctx context.Context, pagination *paginator.Pagination, userID string) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetFavoriteFoldersPaginateService", trace.WithAttributes(attribute.String("service", "GetFavoriteFoldersPaginate")))

	result, err := s.favoriteFolderRepository().GetFavoriteFoldersPaginate(ctx, pagination, userID)

	// Convert the result to response
	if result != nil && result.Data != nil {
		result.Data = new(responses.GetFavoriteFoldersResponse).Make(result.Data.([]models.FavoriteFolder)).FavoriteFolders
	}

	s.tracer.TraceEnd(childSpan)

	return result, err
}

// GetFavoriteFolderById retrieves a favorite folder by ID
func (s Service) GetFavoriteFolderById(ctx context.Context, id int, userID string) (*responses.GetFavoriteFolderByIdResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetFavoriteFolderByIdService", trace.WithAttributes(attribute.String("service", "GetFavoriteFolderById")))

	folder, err := s.favoriteFolderRepository().GetFavoriteFolderById(ctx, id, userID)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return new(responses.GetFavoriteFolderByIdResponse).Make(folder), nil
}

// AddPropertyToFavorite adds a property to a favorite folder (handles 3 cases)
func (s Service) AddPropertyToFavorite(ctx context.Context, dto *dtos.AddPropertyToFavorite, userID string) (*responses.AddPropertyToFavoriteResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "AddPropertyToFavoriteService", trace.WithAttributes(attribute.String("service", "AddPropertyToFavorite")))

	// Validate: Cannot specify both folderId and folderName
	if dto.FolderID != nil && dto.FolderName != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, exception.ErrInvalidRequestParameter
	}

	// Case A: Create empty folder (no propertyId)
	if dto.PropertyID == nil {
		// Must provide folderName when creating empty folder
		if dto.FolderName == nil || *dto.FolderName == "" {
			s.tracer.TraceEnd(childSpan)
			return nil, exception.ErrInvalidRequestParameter
		}
		// Cannot specify folderId without propertyId
		if dto.FolderID != nil {
			s.tracer.TraceEnd(childSpan)
			return nil, exception.ErrInvalidRequestParameter
		}

		newFolder, err := s.favoriteFolderRepository().CreateFavoriteFolder(ctx, *dto.FolderName, userID)
		if err != nil {
			s.tracer.TraceEnd(childSpan)
			return nil, err
		}

		folder, err := s.favoriteFolderRepository().GetFavoriteFolderWithProperties(ctx, newFolder.ID, userID)
		if err != nil {
			s.tracer.TraceEnd(childSpan)
			return nil, err
		}

		s.tracer.TraceEnd(childSpan)
		return new(responses.AddPropertyToFavoriteResponse).Make(folder, true), nil
	}

	// Validate property exists and is active before creating folder
	if err := s.favoriteFolderRepository().ValidateActiveProperty(ctx, *dto.PropertyID); err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	// Determine folder ID to use
	folderID, isNewFolder, err := s.determineFolderID(ctx, dto, userID)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	// Add property to folder
	err = s.favoriteFolderRepository().AddPropertyToFolder(ctx, folderID, *dto.PropertyID, userID)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	// Get folder with properties for response
	folder, err := s.favoriteFolderRepository().GetFavoriteFolderWithProperties(ctx, folderID, userID)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	s.tracer.TraceEnd(childSpan)

	return new(responses.AddPropertyToFavoriteResponse).Make(folder, isNewFolder), nil
}

// determineFolderID determines which folder to use based on DTO (Case 1, 2, or 3)
func (s Service) determineFolderID(ctx context.Context, dto *dtos.AddPropertyToFavorite, userID string) (int, bool, error) {
	// Case 1: Use existing folder
	if dto.FolderID != nil {
		return *dto.FolderID, false, nil
	}

	// Case 2 & 3: Create new folder
	folderName := s.determineFolderName(ctx, dto, userID)
	if folderName == "" {
		return 0, false, exception.ErrInvalidRequestParameter
	}

	newFolder, err := s.favoriteFolderRepository().CreateFavoriteFolder(ctx, folderName, userID)
	if err != nil {
		return 0, false, err
	}

	return newFolder.ID, true, nil
}

// determineFolderName determines the folder name for new folder (Case 2 or 3)
func (s Service) determineFolderName(ctx context.Context, dto *dtos.AddPropertyToFavorite, userID string) string {
	// Case 2: Use specified folder name
	if dto.FolderName != nil && *dto.FolderName != "" {
		return *dto.FolderName
	}

	// Case 3: Generate default name "My Wishlist {n+1}"
	count, err := s.favoriteFolderRepository().GetFavoriteFolderCountByUserId(ctx, userID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get folder count: %v", err))
		return "My Wishlist 1" // fallback to 1 if count fails
	}

	return fmt.Sprintf("My Wishlist %d", count+1)
}

// UpdateFolderName updates the name of a favorite folder
func (s Service) UpdateFolderName(ctx context.Context, folderID int, name string, userID string) (*responses.UpdateFolderNameResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UpdateFolderNameService", trace.WithAttributes(attribute.String("service", "UpdateFolderName")))

	// Update folder name
	folder, err := s.favoriteFolderRepository().UpdateFolderName(ctx, folderID, name, userID)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	s.tracer.TraceEnd(childSpan)

	return new(responses.UpdateFolderNameResponse).Make(folder), nil
}

// DeleteFolder soft deletes a favorite folder
func (s Service) DeleteFolder(ctx context.Context, folderID int, userID string) (*responses.DeleteFolderResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "DeleteFolderService", trace.WithAttributes(attribute.String("service", "DeleteFolder")))

	// Delete folder
	err := s.favoriteFolderRepository().DeleteFolder(ctx, folderID, userID)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	s.tracer.TraceEnd(childSpan)

	return new(responses.DeleteFolderResponse).Make(), nil
}

// RemovePropertyFromAllFolders removes a property from all favorite folders of the user
// Empty folders after removal will be automatically deleted
func (s Service) RemovePropertyFromAllFolders(ctx context.Context, propertyID int, userID string) (*responses.RemovePropertyFromAllFoldersResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "RemovePropertyFromAllFoldersService", trace.WithAttributes(attribute.String("service", "RemovePropertyFromAllFolders")))

	// Remove property from all user's folders
	foldersAffected, foldersDeleted, err := s.favoriteFolderRepository().RemovePropertyFromAllUserFolders(ctx, propertyID, userID)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	s.tracer.TraceEnd(childSpan)

	return new(responses.RemovePropertyFromAllFoldersResponse).Make(foldersAffected, foldersDeleted), nil
}
