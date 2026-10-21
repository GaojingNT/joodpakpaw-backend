package property_module

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/config"
	"joodpakpaw/internal/exception"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/action_log_module"
	"joodpakpaw/src/modules/notification_module"
	"joodpakpaw/src/modules/property_module/dtos"
	"joodpakpaw/src/modules/property_module/responses"
)

func (s Service) GetProperties(ctx context.Context, paginate *paginator.Pagination, userID string) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetPropertiesService", trace.WithAttributes(attribute.String("service", "GetProperties")))

	minRating, _ := paginate.GetFloatAttribute("min_rating")
	sortBy, _ := paginate.GetStringAttribute("sort_by")
	days, _ := paginate.GetIntAttribute("days")

	var result *paginator.Pagination
	var err error

	switch {
	case minRating > 0:
		result, err = s.getTopRatedProperties(ctx, paginate, minRating)
	case sortBy == "price":
		result, err = s.getBestValueProperties(ctx, paginate)
	case days > 0:
		result, err = s.getNewListingsProperties(ctx, paginate, days)
	case sortBy == "trending":
		result, err = s.getTrendingProperties(ctx, paginate)
	default:
		result, err = s.propertyRepository().GetPropertyPaginate(ctx, paginate)
	}

	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	// Get favorited property IDs for the user
	var favoritedPropertyIDs map[int]bool
	if userID != "" {
		favoritedPropertyIDs, _ = s.propertyRepository().GetFavoritedPropertyIDs(ctx, userID)
	}

	// Convert the result to a collection of responses
	if result.Data != nil {
		result.Data = new(responses.GetPropertiesCollection).Collection(result.Data.([]models.Property), favoritedPropertyIDs)
	}

	s.tracer.TraceEnd(childSpan)
	return result, nil
}

func (s Service) getTopRatedProperties(ctx context.Context, paginate *paginator.Pagination, minRating float64) (*paginator.Pagination, error) {
	ids, err := s.propertyRepository().GetTopRatedPropertyIDs(ctx, minRating)
	if err != nil {
		return nil, err
	}
	return s.fetchPropertiesByOrderedIDs(ctx, paginate, ids)
}

func (s Service) getBestValueProperties(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error) {
	ids, err := s.propertyRepository().GetBestValuePropertyIDs(ctx)
	if err != nil {
		return nil, err
	}
	return s.fetchPropertiesByOrderedIDs(ctx, paginate, ids)
}

func (s Service) getNewListingsProperties(ctx context.Context, paginate *paginator.Pagination, days int) (*paginator.Pagination, error) {
	ids, err := s.propertyRepository().GetNewListingsPropertyIDs(ctx, days)
	if err != nil {
		return nil, err
	}

	// Fallback: if results < 5 on page 1, return 20 most recent instead
	if len(ids) < 5 && paginate.GetPage() == 1 {
		ids, err = s.propertyRepository().GetRecentActivePropertyIDs(ctx, 20)
		if err != nil {
			return nil, err
		}
		// Override limit to return all 20 at once on page 1
		paginate.Limit = 20
	}

	return s.fetchPropertiesByOrderedIDs(ctx, paginate, ids)
}

func (s Service) getTrendingProperties(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error) {
	ids, err := s.propertyRepository().GetTrendingPropertyIDs(ctx)
	if err != nil {
		return nil, err
	}
	return s.fetchPropertiesByOrderedIDs(ctx, paginate, ids)
}

// fetchPropertiesByOrderedIDs paginates a pre-ordered list of property IDs in memory,
// then fetches full property data (with preloads) for the current page.
// petTypes/maxPrice filters from paginate attributes are applied before pagination.
func (s Service) fetchPropertiesByOrderedIDs(ctx context.Context, paginate *paginator.Pagination, ids []int) (*paginator.Pagination, error) {
	// Exclude properties owned by the requesting user
	excludeHostIDStr, _ := paginate.GetStringAttribute("exclude_host_id")
	if excludeHostID, err := strconv.Atoi(excludeHostIDStr); err == nil && excludeHostID > 0 {
		hostPropIDs, _ := s.propertyRepository().GetPropertyIDsByHostID(ctx, excludeHostID)
		if len(hostPropIDs) > 0 {
			excludeSet := make(map[int]bool, len(hostPropIDs))
			for _, id := range hostPropIDs {
				excludeSet[id] = true
			}
			filtered := make([]int, 0, len(ids))
			for _, id := range ids {
				if !excludeSet[id] {
					filtered = append(filtered, id)
				}
			}
			ids = filtered
		}
	}

	// Apply petTypes/maxPrice filter if set
	petTypesStr, _ := paginate.GetStringAttribute("pet_types")
	maxPrice, _ := paginate.GetFloatAttribute("max_price")

	if petTypesStr != "" || maxPrice > 0 {
		var petTypes []string
		for _, t := range strings.Split(petTypesStr, ",") {
			if t = strings.TrimSpace(t); t != "" {
				petTypes = append(petTypes, t)
			}
		}
		filterIDs, err := s.propertyRepository().GetPropertyIDsForFilter(ctx, petTypes, maxPrice)
		if err != nil {
			return nil, err
		}
		if filterIDs != nil {
			filtered := make([]int, 0, len(ids))
			for _, id := range ids {
				if filterIDs[id] {
					filtered = append(filtered, id)
				}
			}
			ids = filtered
		}
	}

	pageIDs := paginator.PaginateInMemory(ids, paginate)
	if len(pageIDs) == 0 {
		paginate.Data = []models.Property{}
		return paginate, nil
	}

	properties, err := s.propertyRepository().GetPropertiesByIDs(ctx, pageIDs)
	if err != nil {
		return nil, err
	}

	// Re-order to match the SQL-determined order
	propMap := make(map[int]models.Property, len(properties))
	for _, p := range properties {
		propMap[p.ID] = p
	}
	ordered := make([]models.Property, 0, len(pageIDs))
	for _, id := range pageIDs {
		if p, ok := propMap[id]; ok {
			ordered = append(ordered, p)
		}
	}

	paginate.Data = ordered
	return paginate, nil
}

func (s Service) GetAllProperties(ctx context.Context, params map[string]string, userID string) (*responses.GetAllPropertiesResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetAllPropertiesService", trace.WithAttributes(attribute.String("service", "GetAllProperties")))

	properties, err := s.propertyRepository().GetAllProperties(ctx, params["search"], params["searchBy"])
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	// Get favorited property IDs for the user
	var favoritedPropertyIDs map[int]bool
	if userID != "" {
		favoritedPropertyIDs, _ = s.propertyRepository().GetFavoritedPropertyIDs(ctx, userID)
	}

	// Convert to response
	response := &responses.GetAllPropertiesResponse{
		Properties: new(responses.GetPropertiesCollection).Collection(properties, favoritedPropertyIDs),
	}

	s.tracer.TraceEnd(childSpan)

	return response, nil
}

func (s Service) GetProperty(ctx context.Context, id int) (*responses.GetPropertyByIDResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetPropertyService", trace.WithAttributes(attribute.String("service", "GetProperty")))

	property, err := s.propertyRepository().GetPropertyByID(ctx, id)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	if property.IsBannedByAdmin {
		s.tracer.TraceEnd(childSpan)
		return nil, exception.ErrRecordNotFound
	}

	var hostAvgRating float64
	if property.HostID != nil {
		var ratingErr error
		hostAvgRating, ratingErr = s.propertyRepository().GetHostAvgRating(ctx, *property.HostID)
		if ratingErr != nil {
			s.logger.Error("GetHostAvgRating error: " + ratingErr.Error())
		}
	}

	s.tracer.TraceEnd(childSpan)

	return new(responses.GetPropertyByIDResponse).Make(property, hostAvgRating), nil
}

func (s Service) CreateProperty(ctx context.Context, propertyDto *dtos.CreateProperty) (*models.Property, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "CreatePropertyService", trace.WithAttributes(attribute.String("service", "CreateProperty")))

	property := new(models.Property)
	property.Title = propertyDto.Title
	property.Description = propertyDto.Description
	property.Latitude = propertyDto.Latitude
	property.Longitude = propertyDto.Longitude
	property.Address = propertyDto.Address
	property.Province = propertyDto.Province
	property.District = propertyDto.District
	property.Country = propertyDto.Country
	property.Zipcode = propertyDto.Zipcode
	property.HostID = propertyDto.HostID
	property.PropertyTypeID = propertyDto.PropertyTypeID

	// Load facilities from database if facility IDs are provided
	if len(propertyDto.Facilities) > 0 {
		facilities, err := s.propertyRepository().GetFacilitiesByIDs(ctx, propertyDto.Facilities)
		if err != nil {
			s.tracer.TraceEnd(childSpan)
			return nil, err
		}
		property.Facilities = facilities
	}

	// Create property in database
	err := s.propertyRepository().CreateProperty(ctx, property)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	// Handle image uploads if images are provided
	if len(propertyDto.Images) > 0 {
		propertyImages := make([]models.PropertyImage, 0, len(propertyDto.Images))
		basePath := "properties"

		for index, fileHeader := range propertyDto.Images {
			// Generate unique filename with index
			timestamp := time.Now().UnixNano()
			ext := filepath.Ext(fileHeader.Filename)
			fileName := fmt.Sprintf("%d_%d_%d%s", property.ID, timestamp, index, ext)

			// Upload file to storage (pass FileHeader directly)
			err := s.fileSystem.Put(ctx, basePath, fileName, fileHeader)
			if err != nil {
				s.logger.Error(fmt.Sprintf("Failed to upload file: %v", err))
				continue
			}

			// Get public URL for the uploaded file
			publicURL := s.fileSystem.GetPublicURL(basePath, fileName)

			// Create PropertyImage record
			propertyImage := models.PropertyImage{
				FileName:          fileHeader.Filename,
				FilePath:          publicURL,
				FileContentType:   fileHeader.Header.Get("Content-Type"),
				FileExtensionName: strings.TrimPrefix(ext, "."),
				FileSize:          int(fileHeader.Size),
				PropertyID:        &property.ID,
			}

			propertyImages = append(propertyImages, propertyImage)
		}

		// Save all property images to database
		if len(propertyImages) > 0 {
			err = s.propertyRepository().CreatePropertyImages(ctx, propertyImages)
			if err != nil {
				s.logger.Error(fmt.Sprintf("Failed to save property images: %v", err))
			}
		}
	}

	// Upgrade user to host after successfully creating property
	if propertyDto.HostID != nil {
		err = s.userRepository().UpgradeUserToHost(ctx, *propertyDto.HostID)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to upgrade user to host: %v", err))
			// Don't return error here, property is already created successfully
		}
	}

	s.tracer.TraceEnd(childSpan)

	return property, nil
}

func (s Service) UpdateProperty(ctx context.Context, id int, propertyDto *dtos.UpdateProperty) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UpdatePropertyService", trace.WithAttributes(attribute.String("service", "UpdateProperty"), attribute.Int64("id", int64(id))))

	property := new(models.Property)
	property.Title = propertyDto.Title
	property.Description = propertyDto.Description
	property.Latitude = propertyDto.Latitude
	property.Longitude = propertyDto.Longitude
	property.Address = propertyDto.Address
	property.Province = propertyDto.Province
	property.District = propertyDto.District
	property.Country = propertyDto.Country
	property.Zipcode = propertyDto.Zipcode
	property.HostID = propertyDto.HostID
	property.PropertyTypeID = propertyDto.PropertyTypeID

	// Load facilities from database if facility IDs are provided
	if len(propertyDto.Facilities) > 0 {
		facilities, err := s.propertyRepository().GetFacilitiesByIDs(ctx, propertyDto.Facilities)
		if err != nil {
			s.tracer.TraceEnd(childSpan)
			return err
		}
		property.Facilities = facilities
	}

	// Update property in database
	err := s.propertyRepository().UpdateProperty(ctx, id, property)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}

	// Delete specific images if deleteImageIds are provided
	if len(propertyDto.DeleteImageIds) > 0 {
		for _, imageID := range propertyDto.DeleteImageIds {
			err := s.propertyRepository().DeletePropertyImage(ctx, imageID)
			if err != nil {
				s.logger.Error(fmt.Sprintf("Failed to delete image %d: %v", imageID, err))
				// Continue with other deletions even if one fails
			}
		}
	}

	// Add new images if provided (without deleting old images)
	if len(propertyDto.Images) > 0 {
		propertyImages := make([]models.PropertyImage, 0, len(propertyDto.Images))
		basePath := "properties"

		for index, fileHeader := range propertyDto.Images {
			// Generate unique filename with index
			timestamp := time.Now().UnixNano()
			ext := filepath.Ext(fileHeader.Filename)
			fileName := fmt.Sprintf("%d_%d_%d%s", id, timestamp, index, ext)

			// Upload file to storage (pass FileHeader directly)
			err := s.fileSystem.Put(ctx, basePath, fileName, fileHeader)
			if err != nil {
				s.logger.Error(fmt.Sprintf("Failed to upload file: %v", err))
				continue
			}

			// Get public URL for the uploaded file
			publicURL := s.fileSystem.GetPublicURL(basePath, fileName)

			// Create PropertyImage record
			propertyImage := models.PropertyImage{
				FileName:          fileHeader.Filename,
				FilePath:          publicURL,
				FileContentType:   fileHeader.Header.Get("Content-Type"),
				FileExtensionName: strings.TrimPrefix(ext, "."),
				FileSize:          int(fileHeader.Size),
				PropertyID:        &id,
			}

			propertyImages = append(propertyImages, propertyImage)
		}

		// Save all property images to database
		if len(propertyImages) > 0 {
			err = s.propertyRepository().CreatePropertyImages(ctx, propertyImages)
			if err != nil {
				s.logger.Error(fmt.Sprintf("Failed to save property images: %v", err))
			}
		}
	}

	s.tracer.TraceEnd(childSpan)

	return nil
}

func (s Service) UnbanProperty(ctx context.Context, id int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UnbanPropertyService", trace.WithAttributes(attribute.String("service", "UnbanProperty"), attribute.Int64("id", int64(id))))

	err := s.propertyRepository().UnbanProperty(ctx, id)

	s.tracer.TraceEnd(childSpan)
	return err
}

func (s Service) UpdatePropertyIsActive(ctx context.Context, id int, propertyDto *dtos.UpdatePropertyIsActive) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UpdatePropertyIsActiveService", trace.WithAttributes(attribute.String("service", "UpdatePropertyIsActive"), attribute.Int64("id", int64(id))))

	if *propertyDto.IsActive {
		// Block if property is banned by admin
		property, err := s.propertyRepository().GetPropertyByID(ctx, id)
		if err != nil {
			s.tracer.TraceEnd(childSpan)
			return err
		}
		if property.IsBannedByAdmin {
			s.tracer.TraceEnd(childSpan)
			return exception.ErrForbidden
		}

		// Validate: property must have at least one room category
		roomCount, err := s.propertyRepository().CountRoomCategoriesByPropertyID(ctx, id)
		if err != nil {
			s.tracer.TraceEnd(childSpan)
			return err
		}
		if roomCount == 0 {
			s.tracer.TraceEnd(childSpan)
			return exception.ErrPropertyRequiresRoomCategory
		}
	}

	err := s.propertyRepository().UpdatePropertyIsActive(ctx, id, *propertyDto.IsActive)

	s.tracer.TraceEnd(childSpan)

	return err
}

func (s Service) DeleteProperty(ctx context.Context, id int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "DeletePropertyService", trace.WithAttributes(attribute.String("service", "DeleteProperty")))

	err := s.propertyRepository().DeleteProperty(ctx, id)

	s.tracer.TraceEnd(childSpan)

	return err
}

func (s Service) GetPropertiesForHost(ctx context.Context, hostID int) (*responses.GetPropertiesForHostResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetPropertiesForHostService", trace.WithAttributes(attribute.String("service", "GetPropertiesForHost"), attribute.Int64("hostId", int64(hostID))))

	properties, err := s.propertyRepository().GetPropertiesByHostID(ctx, hostID)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	propertyIDs := make([]int, 0, len(properties))
	for _, p := range properties {
		propertyIDs = append(propertyIDs, p.ID)
	}
	unbanCounts, _ := s.propertyRepository().CountUnbanRequestsByPropertyIDs(ctx, propertyIDs)
	pendingUnban, _ := s.propertyRepository().GetPendingUnbanPropertyIDs(ctx, propertyIDs)

	response := &responses.GetPropertiesForHostResponse{
		Properties: new(responses.PropertiesForHostCollection).Collection(properties, unbanCounts, pendingUnban),
	}

	s.tracer.TraceEnd(childSpan)

	return response, nil
}

func (s Service) GetPropertiesForHostPaginate(ctx context.Context, hostID int, paginate *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetPropertiesForHostPaginateService", trace.WithAttributes(attribute.String("service", "GetPropertiesForHostPaginate"), attribute.Int64("hostId", int64(hostID))))

	result, err := s.propertyRepository().GetPropertiesByHostIDPaginate(ctx, hostID, paginate)

	if result.Data != nil {
		props := result.Data.([]models.Property)
		propertyIDs := make([]int, 0, len(props))
		for _, p := range props {
			propertyIDs = append(propertyIDs, p.ID)
		}
		unbanCounts, _ := s.propertyRepository().CountUnbanRequestsByPropertyIDs(ctx, propertyIDs)
		pendingUnban, _ := s.propertyRepository().GetPendingUnbanPropertyIDs(ctx, propertyIDs)
		result.Data = new(responses.PropertiesForHostCollection).Collection(props, unbanCounts, pendingUnban)
	}

	s.tracer.TraceEnd(childSpan)

	return result, err
}

func (s Service) GetPropertyForHostByID(ctx context.Context, hostID int, propertyID int) (*responses.GetPropertyForHostByIDResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetPropertyForHostByIDService", trace.WithAttributes(attribute.String("service", "GetPropertyForHostByID"), attribute.Int64("hostId", int64(hostID)), attribute.Int64("propertyId", int64(propertyID))))

	property, err := s.propertyRepository().GetPropertyForHostByID(ctx, hostID, propertyID)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	hostAvgRating, ratingErr := s.propertyRepository().GetHostAvgRating(ctx, hostID)
	if ratingErr != nil {
		s.logger.Error("GetHostAvgRating error: " + ratingErr.Error())
	}

	unbanCount, _ := s.propertyRepository().CountUnbanRequestsByPropertyID(ctx, propertyID)
	hasPendingUnban, _ := s.propertyRepository().HasPendingUnbanRequestByPropertyID(ctx, propertyID)

	s.tracer.TraceEnd(childSpan)

	return new(responses.GetPropertyForHostByIDResponse).Make(property, hostAvgRating, unbanCount, hasPendingUnban), nil
}

func (s Service) GetRecommendedProperties(ctx context.Context, userID int, paginate *paginator.Pagination, authUserID string) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetRecommendedPropertiesService", trace.WithAttributes(attribute.String("service", "GetRecommendedProperties"), attribute.Int64("userId", int64(userID))))

	// Call ML service
	mlURL := fmt.Sprintf("%s/recommend/user/%d", config.Global.App.MLServiceURL, userID)
	resp, err := http.Get(mlURL) //nolint:gosec
	if err != nil {
		s.logger.Error(fmt.Sprintf("ML service unavailable: %v", err))
		paginate.Data = new(responses.GetPropertiesCollection).Collection([]models.Property{}, nil)
		s.tracer.TraceEnd(childSpan)
		return paginate, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error(fmt.Sprintf("ML service returned non-200 status: %d", resp.StatusCode))
		paginate.Data = new(responses.GetPropertiesCollection).Collection([]models.Property{}, nil)
		s.tracer.TraceEnd(childSpan)
		return paginate, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to read ML response: %v", err))
		paginate.Data = new(responses.GetPropertiesCollection).Collection([]models.Property{}, nil)
		s.tracer.TraceEnd(childSpan)
		return paginate, nil
	}

	var mlItems []responses.MLRecommendItem
	if err = json.Unmarshal(body, &mlItems); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to parse ML response: %v | body: %s", err, string(body)))
		paginate.Data = new(responses.GetPropertiesCollection).Collection([]models.Property{}, nil)
		s.tracer.TraceEnd(childSpan)
		return paginate, nil
	}

	// Extract property IDs (preserving ML similarity order)
	allIDs := make([]int, 0, len(mlItems))
	for _, item := range mlItems {
		allIDs = append(allIDs, item.PropertyID)
	}

	// Paginate the ID slice using paginator helper
	pageIDs := paginator.PaginateInMemory(allIDs, paginate)
	if len(pageIDs) == 0 {
		paginate.Data = new(responses.GetPropertiesCollection).Collection([]models.Property{}, nil)
		s.tracer.TraceEnd(childSpan)
		return paginate, nil
	}

	// Fetch properties from DB
	properties, err := s.propertyRepository().GetPropertiesByIDs(ctx, pageIDs)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	// Re-order to match ML similarity order
	propMap := make(map[int]models.Property, len(properties))
	for _, p := range properties {
		propMap[p.ID] = p
	}
	ordered := make([]models.Property, 0, len(pageIDs))
	for _, id := range pageIDs {
		if p, ok := propMap[id]; ok {
			ordered = append(ordered, p)
		}
	}

	// Get favorited property IDs for the user
	var favoritedPropertyIDs map[int]bool
	if authUserID != "" {
		favoritedPropertyIDs, _ = s.propertyRepository().GetFavoritedPropertyIDs(ctx, authUserID)
	}

	paginate.Data = new(responses.GetPropertiesCollection).Collection(ordered, favoritedPropertyIDs)

	s.tracer.TraceEnd(childSpan)
	return paginate, nil
}

func (s Service) GetHostReviews(ctx context.Context, hostID int, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetHostReviewsService", trace.WithAttributes(attribute.String("service", "GetHostReviews"), attribute.Int64("hostId", int64(hostID))))
	defer s.tracer.TraceEnd(childSpan)

	result, err := s.propertyRepository().GetHostReviewsPaginate(ctx, hostID, pagination)
	if err != nil {
		return nil, err
	}

	if result.Data != nil {
		result.Data = new(responses.HostReviewsCollection).Collection(result.Data.([]models.Booking))
	}

	return result, nil
}

func (s Service) GetHostProfile(ctx context.Context, hostID int) (*responses.GetHostProfileResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetHostProfileService", trace.WithAttributes(attribute.String("service", "GetHostProfile"), attribute.Int64("hostId", int64(hostID))))
	defer s.tracer.TraceEnd(childSpan)

	host, err := s.userRepository().GetUserByID(ctx, hostID)
	if err != nil {
		return nil, err
	}

	properties, err := s.propertyRepository().GetPropertiesByHostID(ctx, hostID)
	if err != nil {
		return nil, err
	}

	hostAvgRating, _ := s.propertyRepository().GetHostAvgRating(ctx, hostID)

	// Count unique reviews across all properties
	reviewCount := 0
	uniqueBookings := make(map[int]bool)
	for _, prop := range properties {
		for _, order := range prop.BookingOrders {
			if order.Booking != nil && order.Booking.Rating != nil && !uniqueBookings[order.Booking.ID] {
				uniqueBookings[order.Booking.ID] = true
				reviewCount++
			}
		}
	}

	unbanAllCount, _ := s.propertyRepository().CountUnbanAllPropertyRequests(ctx, hostID)
	hasPendingUnbanAll, _ := s.propertyRepository().HasPendingUnbanAllPropertyRequest(ctx, hostID)

	return new(responses.GetHostProfileResponse).Make(host, properties, hostAvgRating, reviewCount, unbanAllCount, hasPendingUnbanAll), nil
}

func (s Service) AddPropertyImages(ctx context.Context, propertyID int, hostID int, dto *dtos.AddPropertyImages) ([]models.PropertyImage, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "AddPropertyImagesService", trace.WithAttributes(attribute.String("service", "AddPropertyImages"), attribute.Int64("propertyId", int64(propertyID))))

	// Verify property belongs to host
	_, err := s.propertyRepository().GetPropertyForHostByID(ctx, hostID, propertyID)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	// Handle image uploads
	propertyImages := make([]models.PropertyImage, 0, len(dto.Images))
	basePath := "properties"

	for index, fileHeader := range dto.Images {
		// Generate unique filename with index
		timestamp := time.Now().UnixNano()
		ext := filepath.Ext(fileHeader.Filename)
		fileName := fmt.Sprintf("%d_%d_%d%s", propertyID, timestamp, index, ext)

		// Upload file to storage
		err := s.fileSystem.Put(ctx, basePath, fileName, fileHeader)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to upload file: %v", err))
			continue
		}

		// Get public URL for the uploaded file
		publicURL := s.fileSystem.GetPublicURL(basePath, fileName)

		// Create PropertyImage record
		propertyImage := models.PropertyImage{
			FileName:          fileHeader.Filename,
			FilePath:          publicURL,
			FileContentType:   fileHeader.Header.Get("Content-Type"),
			FileExtensionName: strings.TrimPrefix(ext, "."),
			FileSize:          int(fileHeader.Size),
			PropertyID:        &propertyID,
		}

		propertyImages = append(propertyImages, propertyImage)
	}

	// Save all property images to database
	if len(propertyImages) > 0 {
		err = s.propertyRepository().CreatePropertyImages(ctx, propertyImages)
		if err != nil {
			s.tracer.TraceEnd(childSpan)
			return nil, err
		}
	}

	s.tracer.TraceEnd(childSpan)
	return propertyImages, nil
}

func (s Service) AdminDeactivateProperty(ctx context.Context, propertyID int, adminID int, adminNote string) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "AdminDeactivatePropertyService", trace.WithAttributes(attribute.String("service", "AdminDeactivateProperty")))
	defer s.tracer.TraceEnd(childSpan)

	// Collect affected guest IDs BEFORE cancelling bookings (for notifications)
	affectedGuestIDs, _ := s.propertyRepository().GetActiveBookingGuestIDsByPropertyIDs(ctx, []int{propertyID})

	// Get property for host notification
	property, err := s.propertyRepository().AdminGetPropertyByID(ctx, propertyID)
	if err != nil {
		return err
	}

	if err = s.propertyRepository().AdminDeactivateProperty(ctx, propertyID); err != nil {
		return err
	}

	// Notify guests whose bookings were cancelled
	for _, guestID := range affectedGuestIDs {
		gID := guestID
		go notification_module.NewService().SendToUser(ctx, gID, "guest", "cancel", fmt.Sprintf("Your booking has been cancelled because the property '%s' was deactivated by admin", property.Title), "", "bookings", nil)
	}

	// Notify host
	if property.HostID != nil {
		go notification_module.NewService().SendToUser(ctx, *property.HostID, "host", "report", fmt.Sprintf("Your property '%s' has been deactivated by admin", property.Title), adminNote, "reports", nil)
	}

	go func() {
		logEntry := &models.ActionLog{
			ModelName: "Property",
			RecordID:  fmt.Sprintf("%d", propertyID),
			Action:    "deactivate_property",
			Messages:  adminNote,
			ActorID:   &adminID,
			TargetID:  property.HostID,
		}
		_ = action_log_module.NewRepository().CreateActionLog(ctx, logEntry)
	}()

	return nil
}

func (s Service) AdminUnbanProperty(ctx context.Context, propertyID int, adminID int, adminNote string) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "AdminUnbanPropertyService", trace.WithAttributes(attribute.String("service", "AdminUnbanProperty")))
	defer s.tracer.TraceEnd(childSpan)

	// Get property for host notification
	property, err := s.propertyRepository().AdminGetPropertyByID(ctx, propertyID)
	if err != nil {
		return err
	}

	if err = s.propertyRepository().UnbanProperty(ctx, propertyID); err != nil {
		return err
	}

	// Notify host
	if property.HostID != nil {
		go notification_module.NewService().SendToUser(ctx, *property.HostID, "host", "report", fmt.Sprintf("Your property '%s' has been unbanned. Re-publish it from your host dashboard.", property.Title), adminNote, "reports", nil)
	}

	go func() {
		logEntry := &models.ActionLog{
			ModelName: "Property",
			RecordID:  fmt.Sprintf("%d", propertyID),
			Action:    "unban_property",
			Messages:  adminNote,
			ActorID:   &adminID,
			TargetID:  property.HostID,
		}
		_ = action_log_module.NewRepository().CreateActionLog(ctx, logEntry)
	}()

	return nil
}

func (s Service) AdminGetProperties(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "AdminGetPropertiesService", trace.WithAttributes(attribute.String("service", "AdminGetProperties")))
	defer s.tracer.TraceEnd(childSpan)

	result, err := s.propertyRepository().AdminGetPropertiesPaginate(ctx, paginate)
	if err != nil {
		return nil, err
	}

	if result.Data != nil {
		result.Data = new(responses.AdminGetPropertiesCollection).Collection(result.Data.([]models.Property))
	}
	return result, nil
}

func (s Service) AdminGetPropertyByID(ctx context.Context, id int) (*responses.AdminGetPropertyByIDResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "AdminGetPropertyByIDService", trace.WithAttributes(attribute.String("service", "AdminGetPropertyByID")))
	defer s.tracer.TraceEnd(childSpan)

	property, err := s.propertyRepository().AdminGetPropertyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return new(responses.AdminGetPropertyByIDResponse).Make(property), nil
}

func (s Service) GetRecommendedPropertiesByProperty(ctx context.Context, propertyID int, limit int, userID string) (*responses.GetAllPropertiesResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetRecommendedPropertiesByPropertyService", trace.WithAttributes(attribute.String("service", "GetRecommendedPropertiesByProperty"), attribute.Int64("propertyId", int64(propertyID))))

	empty := &responses.GetAllPropertiesResponse{Properties: []responses.PropertyListResponse{}}

	mlURL := fmt.Sprintf("%s/recommend/property/%d", config.Global.App.MLServiceURL, propertyID)
	if limit > 0 {
		mlURL = fmt.Sprintf("%s?limit=%d", mlURL, limit)
	}
	resp, err := http.Get(mlURL) //nolint:gosec
	if err != nil {
		s.logger.Error(fmt.Sprintf("ML service unavailable: %v", err))
		s.tracer.TraceEnd(childSpan)
		return empty, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error(fmt.Sprintf("ML service returned non-200 status: %d", resp.StatusCode))
		s.tracer.TraceEnd(childSpan)
		return empty, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to read ML response: %v", err))
		s.tracer.TraceEnd(childSpan)
		return empty, nil
	}

	var mlItems []responses.MLPropertyRecommendItem
	if err = json.Unmarshal(body, &mlItems); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to parse ML response: %v | body: %s", err, string(body)))
		s.tracer.TraceEnd(childSpan)
		return empty, nil
	}

	recommendedIDs := make([]int, 0, len(mlItems))
	for _, item := range mlItems {
		recommendedIDs = append(recommendedIDs, item.PropertyID)
	}

	if len(recommendedIDs) == 0 {
		s.tracer.TraceEnd(childSpan)
		return empty, nil
	}

	properties, err := s.propertyRepository().GetPropertiesByIDs(ctx, recommendedIDs)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	propMap := make(map[int]models.Property, len(properties))
	for _, p := range properties {
		propMap[p.ID] = p
	}
	ordered := make([]models.Property, 0, len(recommendedIDs))
	for _, id := range recommendedIDs {
		if p, ok := propMap[id]; ok {
			ordered = append(ordered, p)
		}
	}

	var favoritedPropertyIDs map[int]bool
	if userID != "" {
		favoritedPropertyIDs, _ = s.propertyRepository().GetFavoritedPropertyIDs(ctx, userID)
	}

	s.tracer.TraceEnd(childSpan)
	return &responses.GetAllPropertiesResponse{
		Properties: new(responses.GetPropertiesCollection).Collection(ordered, favoritedPropertyIDs),
	}, nil
}
