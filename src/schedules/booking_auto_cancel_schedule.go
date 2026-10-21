package schedules

import (
	"fmt"
	"time"

	"joodpakpaw/app/scheduler"
	"joodpakpaw/src/models"
)

func (s SchedulerTask) BookingAutoCancelJob() {
	s.sc.NewJob(s.cronString, func(ctx scheduler.ISchedulerContext) error {
		ctx.Log("booking-auto-cancel", "Starting auto-cancel job for pending bookings older than 24 hours")

		// Find pending bookings created more than 24 hours ago
		expiredTime := time.Now().Add(-24 * time.Hour)
		var bookings []models.Booking
		if err := ctx.DB().
			Preload("BookingOrders").
			Where("booking_status = ? AND created_at <= ?", models.BookingStatusPending, expiredTime).
			Find(&bookings).Error; err != nil {
			ctx.Log("booking-auto-cancel", fmt.Sprintf("Error querying bookings: %v", err))
			return err
		}

		for _, booking := range bookings {
			tx := ctx.DB().Begin()
			if tx.Error != nil {
				ctx.Log("booking-auto-cancel", fmt.Sprintf("Error starting transaction for booking ID %d: %v", booking.ID, tx.Error))
				continue
			}

			// Return rooms to inventory
			rollbackNeeded := false
			for _, order := range booking.BookingOrders {
				for d := order.CheckInDate; d.Before(order.CheckOutDate); d = d.AddDate(0, 0, 1) {
					var calendar models.RoomCategoryCalendar
					err := tx.Where("room_category_id = ? AND date = ?", order.RoomCategoryID, d).
						First(&calendar).Error

					if err == nil {
						calendar.AvailableRooms += order.RoomsBooked
						calendar.BookedRooms -= order.RoomsBooked

						if err := tx.Save(&calendar).Error; err != nil {
							tx.Rollback()
							ctx.Log("booking-auto-cancel", fmt.Sprintf("Error returning rooms for booking ID %d: %v", booking.ID, err))
							rollbackNeeded = true
							break
						}
					}
				}
				if rollbackNeeded {
					break
				}
			}

			if rollbackNeeded {
				continue
			}

			// Update booking status to cancelled
			if err := tx.Model(&booking).Update("booking_status", models.BookingStatusCancelled).Error; err != nil {
				tx.Rollback()
				ctx.Log("booking-auto-cancel", fmt.Sprintf("Error cancelling booking ID %d: %v", booking.ID, err))
				continue
			}

			if err := tx.Commit().Error; err != nil {
				ctx.Log("booking-auto-cancel", fmt.Sprintf("Error committing transaction for booking ID %d: %v", booking.ID, err))
				continue
			}

			ctx.Log("booking-auto-cancel", fmt.Sprintf("Auto-cancelled booking ID %d", booking.ID))
		}

		ctx.Log("booking-auto-cancel", fmt.Sprintf("Auto-cancel job completed. Processed %d bookings", len(bookings)))
		return nil
	})
}
