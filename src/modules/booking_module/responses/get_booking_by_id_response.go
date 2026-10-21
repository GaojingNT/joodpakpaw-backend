package responses

import (
	"joodpakpaw/src/models"
)

type GetBookingByIDResponse struct {
	ID            int                  `json:"id"`
	CheckInDate   string               `json:"checkInDate"`
	CheckOutDate  string               `json:"checkOutDate"`
	NumberOfPets  int                  `json:"numberOfPets"`
	TotalPrice    float64              `json:"totalPrice"`
	BookingStatus models.BookingStatus `json:"bookingStatus"`
	GuestNote     string               `json:"guestNote,omitempty"`
	Rating        *int                 `json:"rating"`
	Comment       string               `json:"comment,omitempty"`
	SuccessedAt   string               `json:"successedAt,omitempty"`
	ConfirmedAt   string               `json:"confirmedAt,omitempty"`
	GuestID       *int                 `json:"guestId"`
	Guest         *UserInfo            `json:"guest,omitempty"`
	BookingOrders []BookingOrderDetail `json:"bookingOrders,omitempty"`
	PetProfiles   []PetProfileInfo     `json:"petProfiles,omitempty"`
	CreatedAt     string               `json:"createdAt"`
	UpdatedAt     string               `json:"updatedAt"`
}

type PropertyInfo struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Address  string `json:"address,omitempty"`
	District string `json:"district,omitempty"`
	Province string `json:"province,omitempty"`
}

type UserInfo struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

type BookingOrderDetail struct {
	ID             int               `json:"id"`
	CheckInDate    string            `json:"checkInDate"`
	CheckOutDate   string            `json:"checkOutDate"`
	RoomsBooked    int               `json:"roomsBooked"`
	PricePerRoom   float64           `json:"pricePerRoom"`
	TotalPrice     float64           `json:"totalPrice"`
	PropertyID     *int              `json:"propertyId"`
	Property       *PropertyInfo     `json:"property,omitempty"`
	RoomCategoryID *int              `json:"roomCategoryId"`
	RoomCategory   *RoomCategoryInfo `json:"roomCategory,omitempty"`
	RoomTypeID     *int              `json:"roomTypeId,omitempty"`
}

type RoomCategoryInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type PetProfileInfo struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Image        string         `json:"image,omitempty"`
	Sex          string         `json:"sex,omitempty"`
	BirthDate    string         `json:"birthDate,omitempty"`
	ContactName  string         `json:"contactName"`
	ContactPhone string         `json:"contactPhone"`
	Note         string         `json:"note,omitempty"`
	BreedName    string         `json:"breedName,omitempty"`
	PetTypeName  string         `json:"petTypeName,omitempty"`
	PetEnergy    *PetEnergyInfo `json:"petEnergy,omitempty"`
	PetSizeName  string         `json:"petSizeName,omitempty"`
}

type PetEnergyInfo struct {
	ID          int    `json:"id"`
	Rating      int    `json:"rating"`
	Description string `json:"description,omitempty"`
}

func mapPetProfiles(petProfiles []models.PetProfile) []PetProfileInfo {
	if len(petProfiles) == 0 {
		return nil
	}
	result := make([]PetProfileInfo, len(petProfiles))
	for i, p := range petProfiles {
		info := PetProfileInfo{
			ID:           p.ID,
			Name:         p.Name,
			Image:        p.Image,
			Sex:          string(p.Sex),
			ContactName:  p.ContactName,
			ContactPhone: p.ContactPhone,
			Note:         p.Note,
		}
		if !p.BirthDate.IsZero() {
			info.BirthDate = p.BirthDate.Format("2006-01-02")
		}
		if p.Breed != nil {
			info.BreedName = p.Breed.Breed
		}
		if p.PetType != nil {
			info.PetTypeName = p.PetType.Name
		}
		if p.PetEnergy != nil {
			info.PetEnergy = &PetEnergyInfo{
				ID:          p.PetEnergy.ID,
				Rating:      p.PetEnergy.Rating,
				Description: p.PetEnergy.Description,
			}
		}
		if p.PetSizeCategory != nil {
			info.PetSizeName = p.PetSizeCategory.Name
		}
		result[i] = info
	}
	return result
}

func (r *GetBookingByIDResponse) Make(booking models.Booking) *GetBookingByIDResponse {
	r.ID = booking.ID
	r.CheckInDate = booking.CheckInDate.Format("2006-01-02")
	r.CheckOutDate = booking.CheckOutDate.Format("2006-01-02")
	r.NumberOfPets = booking.NumberOfPets
	r.TotalPrice = booking.TotalPrice
	r.BookingStatus = booking.BookingStatus
	r.GuestNote = booking.GuestNote
	r.Rating = booking.Rating
	r.Comment = booking.Comment
	r.GuestID = booking.GuestID
	r.CreatedAt = booking.CreatedAt.Format("2006-01-02 15:04:05")
	r.UpdatedAt = booking.UpdatedAt.Format("2006-01-02 15:04:05")

	// Format nullable timestamps
	if booking.SuccessedAt != nil && booking.SuccessedAt.Valid {
		r.SuccessedAt = booking.SuccessedAt.Time.Format("2006-01-02 15:04:05")
	}
	if booking.ConfirmedAt != nil && booking.ConfirmedAt.Valid {
		r.ConfirmedAt = booking.ConfirmedAt.Time.Format("2006-01-02 15:04:05")
	}

	// Map guest info if available
	if booking.Guest != nil {
		r.Guest = &UserInfo{
			ID:        booking.Guest.ID,
			FirstName: booking.Guest.FirstName,
			LastName:  booking.Guest.LastName,
			Email:     booking.Guest.Email,
		}
	}

	// Map pet profiles if available
	r.PetProfiles = mapPetProfiles(booking.PetProfiles)

	// Map booking orders if available
	if len(booking.BookingOrders) > 0 {
		r.BookingOrders = make([]BookingOrderDetail, len(booking.BookingOrders))
		for i, order := range booking.BookingOrders {
			r.BookingOrders[i] = BookingOrderDetail{
				ID:             order.ID,
				CheckInDate:    order.CheckInDate.Format("2006-01-02"),
				CheckOutDate:   order.CheckOutDate.Format("2006-01-02"),
				RoomsBooked:    order.RoomsBooked,
				PricePerRoom:   order.PricePerRoom,
				TotalPrice:     order.TotalPrice,
				PropertyID:     order.PropertyID,
				RoomCategoryID: order.RoomCategoryID,
			}

			// Map property if available
			if order.Property != nil {
				r.BookingOrders[i].Property = &PropertyInfo{
					ID:       order.Property.ID,
					Title:    order.Property.Title,
					Address:  order.Property.Address,
					District: order.Property.District,
					Province: order.Property.Province,
				}
			}

			// Map room category if available
			if order.RoomCategory != nil {
				r.BookingOrders[i].RoomCategory = &RoomCategoryInfo{
					ID:   order.RoomCategory.ID,
					Name: order.RoomCategory.Name,
				}
			}
		}
	}

	return r
}
