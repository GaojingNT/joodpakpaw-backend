package models

type ActionLog struct {
	BaseModel
	// Fields
	ModelName string `json:"modelName" gorm:"column:model_name;type:varchar(255);index;"`
	Messages  string `json:"messages" gorm:"column:messages;type:text;"`
	RecordID  string `json:"recordId" gorm:"column:record_id;type:varchar(255);index;"`
	Action    string `json:"action" gorm:"column:action;type:varchar(50);index;"`

	// Relationships
	ActorID    *int  `json:"actorId" gorm:"column:actor_id;"`
	Actor      *User `json:"actor,omitempty" gorm:"foreignKey:ActorID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	TargetID   *int  `json:"targetId" gorm:"column:target_id;"`
	Target     *User `json:"target,omitempty" gorm:"foreignKey:TargetID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
