package facility_module

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

func (r Repository) GetFacilityPaginate(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetFacilityPaginateRepository", trace.WithAttributes(attribute.String("repository", "GetFacilityPaginate")))
		facilities   = make([]models.Facility, 0)
		err          error
	)

	// Get attributes
	searchAttribute, _ := pagination.GetStringAttribute("search")
	searchByAttribute, _ := pagination.GetStringAttribute("search_by")
	categoryAttribute, _ := pagination.GetStringAttribute("category")

	// Set tracing attributes
	r.tracer.SetAttributes(childSpan, attribute.String("search", searchAttribute))
	r.tracer.SetAttributes(childSpan, attribute.String("search_by", searchByAttribute))
	r.tracer.SetAttributes(childSpan, attribute.String("category", categoryAttribute))

	// Create pagination query
	tx := r.db.
		Joins("LEFT JOIN tbl_facility_categories ON tbl_facilities.category_id = tbl_facility_categories.id").
		Preload("Category")

	utils.Block{
		Try: func() {
			// Build query
			query := tx

			// Apply search across facility name and category name
			if searchAttribute != "" {
				// Combine searchable fields from both tables
				allSearchableFields := append(models.FacilitySearchable(), models.FacilityCategorySearchable()...)
				query = query.Scopes(models.SearchingScope(allSearchableFields, searchAttribute, searchByAttribute))
			}

			// Filter by category name if provided
			if categoryAttribute != "" {
				query = query.Where("tbl_facility_categories.name = ?", categoryAttribute)
			}

			// Select distinct facilities to avoid duplicates from join
			query = query.Distinct("tbl_facilities.*")

			// Execute query with pagination
			if err = query.
				Scopes(paginator.Paginate(facilities, pagination, r.db)).
				Find(&facilities).Error; err != nil {
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
	pagination.Data = facilities

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return pagination, nil
}

func (r Repository) GetAllFacilities(ctx context.Context) ([]models.Facility, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetAllFacilitiesRepository", trace.WithAttributes(attribute.String("repository", "GetAllFacilities")))
		facilities   []models.Facility
		err          error
	)

	utils.Block{
		Try: func() {
			// Get all facilities with category preloaded
			if err = r.db.Preload("Category").Find(&facilities).Error; err != nil {
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

	return facilities, nil
}

func (r Repository) GetFacilityByID(ctx context.Context, id int) (models.Facility, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetFacilityByIDRepository", trace.WithAttributes(attribute.String("repository", "GetFacilityByID"), attribute.Int64("id", int64(id))))
		facility     models.Facility
		err          error
	)

	utils.Block{
		Try: func() {
			// Get facility by ID with category preloaded
			if err = r.db.Preload("Category").First(&facility, id).Error; err != nil {
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
		return facility, err
	}

	return facility, nil
}

