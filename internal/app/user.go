package app

import "time"

type User struct {
	Id          string     `json:"id" gorm:"column:id" bson:"_id" dynamodbav:"id,omitempty" firestore:"id,omitempty"`
	Username    string     `json:"username,omitempty" gorm:"column:username" bson:"username,omitempty" dynamodbav:"username,omitempty" firestore:"username,omitempty"`
	Active      bool       `json:"active" gorm:"column:active" bson:"active" dynamodbav:"active" firestore:"active" validate:"active"`
	Locked      bool       `json:"locked" gorm:"column:locked" bson:"locked" dynamodbav:"locked" firestore:"locked"`
	DateOfBirth *time.Time `json:"dateOfBirth,omitempty" gorm:"column:dateofbirth" bson:"dateOfBirth,omitempty" dynamodbav:"dateOfBirth,omitempty" firestore:"dateOfBirth,omitempty"`
}
