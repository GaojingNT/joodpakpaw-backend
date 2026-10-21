package models

type NotificationAction struct {
	BaseModel

	Name string `json:"name" gorm:"column:name;type:varchar(100);uniqueIndex;not null"`
}
