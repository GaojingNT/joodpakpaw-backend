package report_module

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/action_log_module"
	"joodpakpaw/src/modules/notification_module"
	"joodpakpaw/src/modules/report_module/dtos"
	"joodpakpaw/src/modules/report_module/responses"
)

func (s Service) GetReports(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetReportsService", trace.WithAttributes(attribute.String("service", "GetReports")))
	defer s.tracer.TraceEnd(span)

	result, err := s.reportRepository().GetReportsPaginate(ctx, pagination)
	if err != nil {
		return nil, err
	}

	if result.Data != nil {
		result.Data = new(responses.GetReportsCollection).Collection(result.Data.([]models.Report))
	}

	return result, nil
}

func (s Service) CreateReportReview(ctx context.Context, reporterID int, dto *dtos.CreateReportReview) (*responses.CreateReportResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "CreateReportReviewService", trace.WithAttributes(attribute.String("service", "CreateReportReview")))
	defer s.tracer.TraceEnd(span)

	category, err := s.reportRepository().GetReportCategoryByID(ctx, *dto.ReportCategoryID)
	if err != nil {
		return nil, err
	}
	if category.Type != models.ReportTypeReview {
		return nil, exception.ErrInvalidRequestParameter
	}

	booking, err := s.reportRepository().GetBookingByID(ctx, *dto.BookingID)
	if err != nil {
		return nil, err
	}
	if booking.GuestID != nil && *booking.GuestID == reporterID {
		return nil, exception.ErrCannotReportOwnReview
	}

	report := &models.Report{
		ReportCategoryID: dto.ReportCategoryID,
		ReporterID:       &reporterID,
		BookingID:        dto.BookingID,
		Description:      dto.Description,
	}

	if err = s.reportRepository().CreateReport(ctx, report); err != nil {
		return nil, err
	}

	s.uploadReportImages(ctx, report.ID, dto.Images)

	go notification_module.NewService().SendToUser(ctx, reporterID, "guest", "report", "Your report on a review has been submitted and is under review by our team", "", "reports", &report.ID)

	return new(responses.CreateReportResponse).Make(*report, models.ReportTypeReview), nil
}

func (s Service) CreateReportProperty(ctx context.Context, reporterID int, dto *dtos.CreateReportProperty) (*responses.CreateReportResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "CreateReportPropertyService", trace.WithAttributes(attribute.String("service", "CreateReportProperty")))
	defer s.tracer.TraceEnd(span)

	category, err := s.reportRepository().GetReportCategoryByID(ctx, *dto.ReportCategoryID)
	if err != nil {
		return nil, err
	}
	if category.Type != models.ReportTypeProperty {
		return nil, exception.ErrInvalidRequestParameter
	}

	property, err := s.reportRepository().GetPropertyByID(ctx, *dto.PropertyID)
	if err != nil {
		return nil, err
	}
	if property.HostID != nil && *property.HostID == reporterID {
		return nil, exception.ErrCannotReportOwnProperty
	}

	report := &models.Report{
		ReportCategoryID: dto.ReportCategoryID,
		ReporterID:       &reporterID,
		PropertyID:       dto.PropertyID,
		Description:      dto.Description,
	}

	if err = s.reportRepository().CreateReport(ctx, report); err != nil {
		return nil, err
	}

	s.uploadReportImages(ctx, report.ID, dto.Images)

	go notification_module.NewService().SendToUser(ctx, reporterID, "guest", "report", "Your report on a property has been submitted and is under review by our team", "", "reports", &report.ID)

	return new(responses.CreateReportResponse).Make(*report, models.ReportTypeProperty), nil
}

func (s Service) CreateReportUnbanProperty(ctx context.Context, reporterID int, dto *dtos.CreateReportUnbanProperty) (*responses.CreateReportResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "CreateReportUnbanPropertyService", trace.WithAttributes(attribute.String("service", "CreateReportUnbanProperty")))
	defer s.tracer.TraceEnd(span)

	category, err := s.reportRepository().GetReportCategoryByID(ctx, *dto.ReportCategoryID)
	if err != nil {
		return nil, err
	}
	if category.Type != models.ReportTypeUnbanProperty {
		return nil, exception.ErrInvalidRequestParameter
	}

	// If requesting single property unban, check ownership and ban status
	if dto.PropertyID != nil {
		property, err := s.reportRepository().GetPropertyByID(ctx, *dto.PropertyID)
		if err != nil {
			return nil, err
		}
		if property.HostID == nil || *property.HostID != reporterID {
			return nil, exception.ErrInvalidRequestParameter
		}

		user, err := s.reportRepository().GetUserByID(ctx, reporterID)
		if err != nil {
			return nil, err
		}
		if user.IsBannedAllProperties {
			return nil, exception.ErrMustUnbanAllProperties
		}
	}

	totalCount, err := s.reportRepository().CountUnbanPropertyRequests(ctx, reporterID, dto.PropertyID)
	if err != nil {
		return nil, err
	}
	if totalCount >= 2 {
		return nil, exception.ErrUnbanRequestLimitExceeded
	}

	hasPending, err := s.reportRepository().HasPendingUnbanRequest(ctx, reporterID, models.ReportTypeUnbanProperty, dto.PropertyID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, exception.ErrUnbanRequestAlreadyPending
	}

	report := &models.Report{
		ReportCategoryID: dto.ReportCategoryID,
		ReporterID:       &reporterID,
		PropertyID:       dto.PropertyID,
		Description:      dto.Description,
	}

	if err = s.reportRepository().CreateReport(ctx, report); err != nil {
		return nil, err
	}

	s.uploadReportImages(ctx, report.ID, dto.Images)

	go notification_module.NewService().SendToUser(ctx, reporterID, "host", "report", "Your unban property request has been submitted and is under review by our team", "", "reports", &report.ID)

	return new(responses.CreateReportResponse).Make(*report, models.ReportTypeUnbanProperty), nil
}

func (s Service) CreateReportUnbanReview(ctx context.Context, reporterID int, dto *dtos.CreateReportUnbanReview) (*responses.CreateReportResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "CreateReportUnbanReviewService", trace.WithAttributes(attribute.String("service", "CreateReportUnbanReview")))
	defer s.tracer.TraceEnd(span)

	category, err := s.reportRepository().GetReportCategoryByID(ctx, *dto.ReportCategoryID)
	if err != nil {
		return nil, err
	}
	if category.Type != models.ReportTypeUnbanReview {
		return nil, exception.ErrInvalidRequestParameter
	}

	hasPending, err := s.reportRepository().HasPendingUnbanRequest(ctx, reporterID, models.ReportTypeUnbanReview, nil)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, exception.ErrUnbanRequestAlreadyPending
	}

	report := &models.Report{
		ReportCategoryID: dto.ReportCategoryID,
		ReporterID:       &reporterID,
		Description:      dto.Description,
	}

	if err = s.reportRepository().CreateReport(ctx, report); err != nil {
		return nil, err
	}

	s.uploadReportImages(ctx, report.ID, dto.Images)

	go notification_module.NewService().SendToUser(ctx, reporterID, "guest", "report", "Your unban review request has been submitted and is under review by our team", "", "reports", &report.ID)

	return new(responses.CreateReportResponse).Make(*report, models.ReportTypeUnbanReview), nil
}

func (s Service) uploadReportImages(ctx context.Context, reportID int, files []*multipart.FileHeader) {
	if len(files) == 0 {
		return
	}

	basePath := "reports"
	reportImages := make([]models.ReportImage, 0, len(files))

	for i, fileHeader := range files {
		ext := filepath.Ext(fileHeader.Filename)
		fileName := fmt.Sprintf("report_%d_%d_%d%s", reportID, time.Now().UnixNano(), i, ext)

		if err := s.fileSystem.Put(ctx, basePath, fileName, fileHeader); err != nil {
			s.logger.Error(fmt.Sprintf("Failed to upload report image: %v", err))
			continue
		}

		publicURL := s.fileSystem.GetPublicURL(basePath, fileName)

		reportImages = append(reportImages, models.ReportImage{
			FileName:          fileHeader.Filename,
			FilePath:          publicURL,
			FileContentType:   fileHeader.Header.Get("Content-Type"),
			FileExtensionName: strings.TrimPrefix(ext, "."),
			FileSize:          int(fileHeader.Size),
			ReportID:          &reportID,
		})
	}

	if len(reportImages) > 0 {
		if err := s.reportRepository().CreateReportImages(ctx, reportImages); err != nil {
			s.logger.Error(fmt.Sprintf("Failed to save report images: %v", err))
		}
	}
}

func (s Service) ActionReport(ctx context.Context, id int, action string, adminNote string, adminID int) error {
	ctx, span := s.tracer.TraceStart(ctx, "ActionReportService", trace.WithAttributes(attribute.String("service", "ActionReport"), attribute.Int64("id", int64(id))))
	defer s.tracer.TraceEnd(span)

	report, err := s.reportRepository().GetReportByID(ctx, id)
	if err != nil {
		return err
	}

	// Cannot action an already actioned report
	if report.Status != models.ReportStatusPending {
		return exception.ErrReportAlreadyActioned
	}

	// Validate action is allowed for report type
	resolvedAction := models.ReportAction(action)
	if report.ReportCategory != nil {
		switch report.ReportCategory.Type {
		case models.ReportTypeProperty:
			if resolvedAction == models.ReportActionDeleteReview || resolvedAction == models.ReportActionBanUserReview {
				return exception.ErrReportActionNotAllowed
			}
		case models.ReportTypeUnbanProperty:
			if resolvedAction != models.ReportActionUnbanProperty && resolvedAction != models.ReportActionUnbanAllProperties && resolvedAction != models.ReportActionDismissReport {
				return exception.ErrReportActionNotAllowed
			}
		case models.ReportTypeUnbanReview:
			if resolvedAction != models.ReportActionUnbanReview && resolvedAction != models.ReportActionDismissReport {
				return exception.ErrReportActionNotAllowed
			}
		case models.ReportTypeReview:
			if resolvedAction == models.ReportActionDeactivateProperty || resolvedAction == models.ReportActionWarnUser {
				return exception.ErrReportActionNotAllowed
			}
		}
	}

	// Collect affected guest IDs BEFORE cancelling bookings (for notifications)
	var affectedGuestIDs []int
	if resolvedAction == models.ReportActionDeactivateProperty && report.PropertyID != nil {
		affectedGuestIDs, _ = s.reportRepository().GetActiveBookingGuestIDsByPropertyIDs(ctx, []int{*report.PropertyID})
	} else if resolvedAction == models.ReportActionDeactivateProperties && report.Property != nil && report.Property.HostID != nil {
		if propertyIDs, err2 := s.reportRepository().GetPropertyIDsByHostID(ctx, *report.Property.HostID); err2 == nil && len(propertyIDs) > 0 {
			affectedGuestIDs, _ = s.reportRepository().GetActiveBookingGuestIDsByPropertyIDs(ctx, propertyIDs)
		}
	}

	if err = s.reportRepository().ActionReport(ctx, report, action, adminNote); err != nil {
		return err
	}

	// Notify guests whose bookings were cancelled
	for _, guestID := range affectedGuestIDs {
		gID := guestID
		go notification_module.NewService().SendToUser(ctx, gID, "guest", "cancel", "Your booking has been cancelled because the property was deactivated by admin", "", "bookings", nil)
	}

	propertyTitle := s.getReportPropertyTitle(report)

	if resolvedAction == models.ReportActionDismissReport && report.ReporterID != nil {
		var msg string
		dismissNoteMsg := ""
		if report.ReportCategory != nil {
			switch report.ReportCategory.Type {
			case models.ReportTypeUnbanProperty:
				msg = "Your account unban request has been reviewed and dismissed by our team."
				dismissNoteMsg = adminNote
			case models.ReportTypeProperty:
				msg = fmt.Sprintf("Your report on '%s' has been reviewed and dismissed by our team", propertyTitle)
				dismissNoteMsg = adminNote
			case models.ReportTypeReview:
				msg = fmt.Sprintf("Your report on '%s' has been reviewed and dismissed by our team", propertyTitle)
			default:
				msg = "Your report has been reviewed and dismissed by our team"
			}
		} else {
			msg = "Your report has been reviewed and dismissed by our team"
		}
		go notification_module.NewService().SendToUser(ctx, *report.ReporterID, "guest", "report", msg, dismissNoteMsg, "reports", &id)
	}

	if resolvedAction == models.ReportActionDeleteReview {
		if report.ReporterID != nil {
			go notification_module.NewService().SendToUser(ctx, *report.ReporterID, "guest", "report", fmt.Sprintf("Your report on '%s' has been reviewed and the review has been removed", propertyTitle), "", "reports", &id)
		}
		if report.Booking != nil && report.Booking.GuestID != nil {
			go notification_module.NewService().SendToUser(ctx, *report.Booking.GuestID, "guest", "report", fmt.Sprintf("Your review on '%s' has been removed due to a violation report", propertyTitle), "", "reports", &id)
		}
	}

	if resolvedAction == models.ReportActionDeactivateProperty {
		if report.ReporterID != nil {
			go notification_module.NewService().SendToUser(ctx, *report.ReporterID, "guest", "report", fmt.Sprintf("Your report on '%s' has been reviewed and the property has been deactivated", propertyTitle), "", "reports", &id)
		}
		if report.Property != nil && report.Property.HostID != nil {
			go notification_module.NewService().SendToUser(ctx, *report.Property.HostID, "host", "report", fmt.Sprintf("Your property '%s' has been deactivated due to a violation report", propertyTitle), adminNote, "reports", &id)
		}
	}

	if resolvedAction == models.ReportActionWarnUser && report.ReportCategory != nil && report.ReportCategory.Type == models.ReportTypeProperty {
		if report.ReporterID != nil {
			go notification_module.NewService().SendToUser(ctx, *report.ReporterID, "guest", "report", fmt.Sprintf("Your report on '%s' has been reviewed and action has been taken", propertyTitle), "", "reports", &id)
		}
		if report.Property != nil && report.Property.HostID != nil {
			go notification_module.NewService().SendToUser(ctx, *report.Property.HostID, "host", "report", fmt.Sprintf("Your property '%s' has received a warning due to a report submitted by a user", propertyTitle), adminNote, "reports", &id)
		}
	}

	if resolvedAction == models.ReportActionBanUserReview {
		if report.ReporterID != nil {
			go notification_module.NewService().SendToUser(ctx, *report.ReporterID, "guest", "report", fmt.Sprintf("Your report on '%s' has been reviewed and the user has been banned from writing reviews", propertyTitle), "", "reports", &id)
		}
		if report.Booking != nil && report.Booking.GuestID != nil {
			go notification_module.NewService().SendToUser(ctx, *report.Booking.GuestID, "guest", "report", fmt.Sprintf("You have been banned from writing reviews due to a violation report on '%s'", propertyTitle), adminNote, "reports", &id)
		}
	}

	if resolvedAction == models.ReportActionDeactivateProperties {
		if report.ReporterID != nil {
			go notification_module.NewService().SendToUser(ctx, *report.ReporterID, "guest", "report", "Your report has been reviewed and all properties of the host have been deactivated", "", "reports", &id)
		}
		if report.Property != nil && report.Property.HostID != nil {
			go notification_module.NewService().SendToUser(ctx, *report.Property.HostID, "host", "report", "All your properties have been deactivated due to a violation report", adminNote, "reports", &id)
			go notification_module.NewService().SendToUser(ctx, *report.Property.HostID, "host", "report", "Your host verification has been revoked", "", "reports", &id)
		}
	}

	if resolvedAction == models.ReportActionUnbanProperty {
		if report.ReporterID != nil {
			go notification_module.NewService().SendToUser(ctx, *report.ReporterID, "host", "report", fmt.Sprintf("Your property '%s' has been unbanned. Re-publish it from your host dashboard.", propertyTitle), adminNote, "reports", &id)
		}
	}

	if resolvedAction == models.ReportActionUnbanAllProperties {
		if report.ReporterID != nil {
			go notification_module.NewService().SendToUser(ctx, *report.ReporterID, "host", "report", "Your account has been reinstated. Re-publish your properties from the host dashboard.", adminNote, "reports", &id)
		}
	}

	if resolvedAction == models.ReportActionUnbanReview {
		if report.ReporterID != nil {
			go notification_module.NewService().SendToUser(ctx, *report.ReporterID, "guest", "report", "Your review ban has been lifted and you can write reviews again", adminNote, "reports", &id)
		}
	}

	go func() {
		logEntry := &models.ActionLog{
			ModelName: "Report",
			RecordID:  fmt.Sprintf("%d", id),
			Action:    action,
			Messages:  adminNote,
			ActorID:   &adminID,
			TargetID:  report.ReporterID,
		}
		_ = action_log_module.NewRepository().CreateActionLog(ctx, logEntry)
	}()

	return nil
}

func (s Service) getReportPropertyTitle(report models.Report) string {
	if report.Property != nil {
		return report.Property.Title
	}
	if report.Booking != nil && len(report.Booking.BookingOrders) > 0 && report.Booking.BookingOrders[0].Property != nil {
		return report.Booking.BookingOrders[0].Property.Title
	}
	return ""
}

func (s Service) GetMyReports(ctx context.Context, userID int, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetMyReportsService", trace.WithAttributes(attribute.String("service", "GetMyReports"), attribute.Int64("userID", int64(userID))))
	defer s.tracer.TraceEnd(span)

	result, err := s.reportRepository().GetMyReportsPaginate(ctx, userID, pagination)
	if err != nil {
		return nil, err
	}

	if result.Data != nil {
		result.Data = new(responses.GetReportsCollection).Collection(result.Data.([]models.Report))
	}

	return result, nil
}

func (s Service) GetMyReportByID(ctx context.Context, id int, userID int) (*responses.GetReportByIDResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetMyReportByIDService", trace.WithAttributes(attribute.String("service", "GetMyReportByID"), attribute.Int64("id", int64(id)), attribute.Int64("userID", int64(userID))))
	defer s.tracer.TraceEnd(span)

	report, err := s.reportRepository().GetMyReportByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	return new(responses.GetReportByIDResponse).Make(report), nil
}

func (s Service) GetReportByID(ctx context.Context, id int) (*responses.GetReportByIDResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetReportByIDService", trace.WithAttributes(attribute.String("service", "GetReportByID"), attribute.Int64("id", int64(id))))
	defer s.tracer.TraceEnd(span)

	report, err := s.reportRepository().GetReportByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return new(responses.GetReportByIDResponse).Make(report), nil
}
