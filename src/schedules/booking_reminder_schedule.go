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

func (s SchedulerTask) BookingReminderJob() {
	s.sc.NewJob(s.cronString, func(ctx scheduler.ISchedulerContext) error {
		ctx.Log("booking-reminder", "Starting check-in reminder job")

		tomorrow := time.Now().AddDate(0, 0, 1)
		tomorrowStart := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())
		tomorrowEnd := tomorrowStart.AddDate(0, 0, 1)

		var bookings []models.Booking
		if err := ctx.DB().
			Preload("BookingOrders.Property").
			Where("booking_status = ? AND check_in_date >= ? AND check_in_date < ?",
				models.BookingStatusConfirmed, tomorrowStart, tomorrowEnd).
			Find(&bookings).Error; err != nil {
			ctx.Log("booking-reminder", fmt.Sprintf("Error querying bookings: %v", err))
			return err
		}

		bgCtx := context.Background()

		frontendURL := os.Getenv("APP_FRONTEND_URL")

		for _, booking := range bookings {
			if len(booking.BookingOrders) == 0 {
				continue
			}

			order := booking.BookingOrders[0]
			if order.Property == nil {
				continue
			}

			propertyName := order.Property.Title
			checkIn := booking.CheckInDate.Format("2 Jan 06")

			// Remind guest
			if booking.GuestID != nil {
				guestTitle := fmt.Sprintf("Your pets' stay at %s starts tomorrow!", propertyName)
				notification_module.NewService().SendToUser(bgCtx, *booking.GuestID, "guest", "reminder", guestTitle, "", "bookings", &booking.ID)
				ctx.Log("booking-reminder", fmt.Sprintf("Sent guest reminder for booking ID %d | check-in: %s", booking.ID, checkIn))

				// Send reminder email to guest
				var guest models.User
				if err := ctx.DB().First(&guest, *booking.GuestID).Error; err == nil {
					guestURL := fmt.Sprintf("%s/profile/bookings/%d/?from=notifications", frontendURL, booking.ID)
					guestName := fmt.Sprintf("%s %s", guest.FirstName, guest.LastName)
					if err := s.mailService.SendReminderEmail(guest.Email, guestName, propertyName, checkIn, booking.NumberOfPets, guestURL, false); err != nil {
						ctx.Log("booking-reminder", fmt.Sprintf("Failed to send reminder email to guest %s: %v", guest.Email, err))
					}
				}
			}

			// Remind host
			if order.Property.HostID != nil {
				hostTitle := fmt.Sprintf("Guest check-in tomorrow at %s | %d pets", propertyName, booking.NumberOfPets)
				notification_module.NewService().SendToUser(bgCtx, *order.Property.HostID, "host", "reminder", hostTitle, "", "bookings", &booking.ID)
				ctx.Log("booking-reminder", fmt.Sprintf("Sent host reminder for booking ID %d | check-in: %s", booking.ID, checkIn))

				// Send reminder email to host
				var host models.User
				if err := ctx.DB().First(&host, *order.Property.HostID).Error; err == nil {
					hostURL := fmt.Sprintf("%s/profile/host/booking/%d/?from=notifications", frontendURL, booking.ID)
					hostName := fmt.Sprintf("%s %s", host.FirstName, host.LastName)
					if err := s.mailService.SendReminderEmail(host.Email, hostName, propertyName, checkIn, booking.NumberOfPets, hostURL, true); err != nil {
						ctx.Log("booking-reminder", fmt.Sprintf("Failed to send reminder email to host %s: %v", host.Email, err))
					}
				}
			}
		}

		ctx.Log("booking-reminder", fmt.Sprintf("Reminder job completed. Processed %d bookings", len(bookings)))
		return nil
	})
}
