package models

type UserBehaviorCategory struct {
	BaseModel

	// Fields
	Name string `json:"name" gorm:"column:name;type:varchar(100);uniqueIndex;not null;"`

	// Relations
	UserBehaviors []UserBehavior `json:"userBehaviors,omitempty" gorm:"foreignKey:BehaviorCategoryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type UserBehavior struct {
	BaseModel

	// Fields
	ReferenceTable  string `json:"referenceTable" gorm:"column:reference_table;type:varchar(100);not null;"`
	ReferenceRecord string `json:"referenceRecord" gorm:"column:reference_record;type:varchar(255);not null;"`

	// Relations
	UserID             int                   `json:"userId" gorm:"column:user_id;not null;index;"`
	User               *User                 `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BehaviorCategoryID int                   `json:"behaviorCategoryId" gorm:"column:behavior_category_id;not null;index;"`
	BehaviorCategory   *UserBehaviorCategory `json:"behaviorCategory,omitempty" gorm:"foreignKey:BehaviorCategoryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
