package schedules

import (
	"context"
	"fmt"
	"os"
	"time"

	"joodpakpaw/app/scheduler"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/notification_module"
)

func (s SchedulerTask) BookingAutoSuccessJob() {
	s.sc.NewJob(s.cronString, func(ctx scheduler.ISchedulerContext) error {
		ctx.Log("booking-auto-success", "Starting auto-success job for confirmed bookings")

		var bookings []models.Booking
		if err := ctx.DB().
			Preload("BookingOrders.Property").
			Where("booking_status = ? AND check_out_date <= CURRENT_DATE", models.BookingStatusConfirmed).
			Find(&bookings).Error; err != nil {
			ctx.Log("booking-auto-success", fmt.Sprintf("Error querying bookings: %v", err))
			return err
		}

		now := time.Now()
		bgCtx := context.Background()
		frontendURL := os.Getenv("APP_FRONTEND_URL")

		for _, booking := range bookings {
			if err := ctx.DB().Model(&booking).Updates(map[string]interface{}{
				"booking_status": models.BookingStatusSuccessful,
				"successed_at":   now,
			}).Error; err != nil {
				ctx.Log("booking-auto-success", fmt.Sprintf("Error updating booking ID %d: %v", booking.ID, err))
				continue
			}
			ctx.Log("booking-auto-success", fmt.Sprintf("Auto-successful booking ID %d", booking.ID))

			if len(booking.BookingOrders) == 0 || booking.BookingOrders[0].Property == nil {
				continue
			}

			order := booking.BookingOrders[0]
			propertyName := order.Property.Title

			// Notify + email guest
			if booking.GuestID != nil {
				guestTitle := fmt.Sprintf("Your stay at %s is complete!", propertyName)
				notification_module.NewService().SendToUser(bgCtx, *booking.GuestID, "guest", "booking_success", guestTitle, "", "bookings", &booking.ID)

				var guest models.User
				if err := ctx.DB().First(&guest, *booking.GuestID).Error; err == nil {
					guestURL := fmt.Sprintf("%s/profile/bookings/%d/?from=notifications", frontendURL, booking.ID)
					guestName := fmt.Sprintf("%s %s", guest.FirstName, guest.LastName)
					if err := s.mailService.SendStayCompletedEmail(guest.Email, guestName, propertyName, guestURL, false); err != nil {
						ctx.Log("booking-auto-success", fmt.Sprintf("Failed to send email to guest %s: %v", guest.Email, err))
					}
				}
			}

			// Notify + email host
			if order.Property.HostID != nil {
				hostTitle := fmt.Sprintf("Check-out at %s has been completed", propertyName)
				notification_module.NewService().SendToUser(bgCtx, *order.Property.HostID, "host", "booking_success", hostTitle, "", "bookings", &booking.ID)

				var host models.User
				if err := ctx.DB().First(&host, *order.Property.HostID).Error; err == nil {
					hostURL := fmt.Sprintf("%s/profile/host/booking/%d/?from=notifications", frontendURL, booking.ID)
					hostName := fmt.Sprintf("%s %s", host.FirstName, host.LastName)
					if err := s.mailService.SendStayCompletedEmail(host.Email, hostName, propertyName, hostURL, true); err != nil {
						ctx.Log("booking-auto-success", fmt.Sprintf("Failed to send email to host %s: %v", host.Email, err))
					}
				}
			}
		}

		ctx.Log("booking-auto-success", fmt.Sprintf("Auto-success job completed. Processed %d bookings", len(bookings)))
		return nil
	})
}
