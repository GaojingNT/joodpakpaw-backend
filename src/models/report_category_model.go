package models

type ReportType string

const (
	ReportTypeReview   ReportType = "review"
	ReportTypeProperty ReportType = "property"
	ReportTypeUnbanProperty ReportType = "unban_property"
	ReportTypeUnbanReview ReportType = "unban_review"
)

type ReportCategory struct {
	BaseModel

	// Fields
	Name string     `json:"name" gorm:"column:name;size:255;not null;"`
	Label string    `json:"label" gorm:"column:label;size:255;default:'';"`
	Type ReportType `json:"type" gorm:"column:type;type:report_type;not null;"`

	// Relationships
	Reports []Report `json:"reports,omitempty" gorm:"foreignKey:ReportCategoryID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}
