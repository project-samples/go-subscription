package app

import "time"

type User struct {
	Id          string     `json:"id" gorm:"column:id" bson:"_id" dynamodbav:"id,omitempty" firestore:"id,omitempty"`
	Username    string     `json:"username,omitempty" gorm:"column:username" bson:"username,omitempty" dynamodbav:"username,omitempty" firestore:"username,omitempty"`
	Active      bool       `json:"active" gorm:"column:active" bson:"active,omitempty" dynamodbav:"active,omitempty" firestore:"active,omitempty" validate:"active"`
	Locked      bool       `json:"locked" gorm:"column:locked" bson:"locked,omitempty" dynamodbav:"locked,omitempty" firestore:"locked,omitempty"`
	DateOfBirth *time.Time `json:"dateOfBirth,omitempty" gorm:"column:dateofbirth" bson:"dateOfBirth,omitempty" dynamodbav:"dateOfBirth,omitempty" firestore:"dateOfBirth,omitempty"`
}
