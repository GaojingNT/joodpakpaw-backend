package dtos

import "mime/multipart"

type (
	CreateProperty struct {
		Title            string                  `json:"title" form:"title" query:"title" validate:"required,max=255"`
		Description      string                  `json:"description" form:"description" query:"description" validate:"omitempty"`
		Latitude         float64                 `json:"latitude" form:"latitude" query:"latitude" validate:"required,min=-90,max=90"`
		Longitude        float64                 `json:"longitude" form:"longitude" query:"longitude" validate:"required,min=-180,max=180"`
		Address          string                  `json:"address" form:"address" query:"address" validate:"omitempty,max=255"`
		Province         string                  `json:"province" form:"province" query:"province" validate:"omitempty,max=100"`
		District         string                  `json:"district" form:"district" query:"district" validate:"omitempty,max=100"`
		Country          string                  `json:"country" form:"country" query:"country" validate:"omitempty,max=100"`
		Zipcode          string                  `json:"zipcode" form:"zipcode" query:"zipcode" validate:"omitempty,max=10"`
		HostID           *int                    `json:"hostId" validate:"required"`
		PropertyTypeID   *int                    `json:"propertyTypeId" form:"propertyTypeId" query:"propertyTypeId" validate:"required"`
		Images           []*multipart.FileHeader `json:"images" form:"images" validate:"omitempty,dive"`
		FacilitiesString string                  `json:"-" form:"facilities" query:"facilities" validate:"omitempty"`
		Facilities       []int                   `json:"facilities" validate:"omitempty"`
	}
	UpdateProperty struct {
		Title                 string                  `json:"title" form:"title" query:"title" validate:"omitempty,max=255"`
		Description           string                  `json:"description" form:"description" query:"description" validate:"omitempty"`
		Latitude              float64                 `json:"latitude" form:"latitude" query:"latitude" validate:"omitempty"`
		Longitude             float64                 `json:"longitude" form:"longitude" query:"longitude" validate:"omitempty"`
		Address               string                  `json:"address" form:"address" query:"address" validate:"omitempty"`
		Province              string                  `json:"province" form:"province" query:"province" validate:"omitempty,max=100"`
		District              string                  `json:"district" form:"district" query:"district" validate:"omitempty,max=100"`
		Country               string                  `json:"country" form:"country" query:"country" validate:"omitempty,max=100"`
		Zipcode               string                  `json:"zipcode" form:"zipcode" query:"zipcode" validate:"omitempty,max=10"`
		HostID                *int                    `json:"hostId" validate:"omitempty"`
		PropertyTypeID        *int                    `json:"propertyTypeId" form:"propertyTypeId" query:"propertyTypeId" validate:"omitempty"`
		Images                []*multipart.FileHeader `json:"images" form:"images" validate:"omitempty,dive"`
		DeleteImageIdsString  string                  `json:"-" form:"deleteImageIds" query:"deleteImageIds" validate:"omitempty"`
		DeleteImageIds        []int                   `json:"deleteImageIds" validate:"omitempty"`
		FacilitiesString      string                  `json:"-" form:"facilities" query:"facilities" validate:"omitempty"`
		Facilities            []int                   `json:"facilities" validate:"omitempty"`
	}

	UpdatePropertyIsActive struct {
		IsActive *bool `json:"isActive" form:"isActive" validate:"required"`
	}

	DeletePropertyImages struct {
		ImageIDs []int `json:"imageIds" validate:"required,min=1,dive,required"`
	}

	AddPropertyImages struct {
		Images []*multipart.FileHeader `form:"images" validate:"required,min=1,dive"`
	}
)
