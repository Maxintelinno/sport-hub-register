package model

import "time"

type User struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Username   string    `json:"username" gorm:"column:username"`
	Fullname   string    `json:"fullname" gorm:"column:fullname"`
	UserType   string    `json:"usertype" gorm:"column:usertype"`
	Password   string    `json:"password" gorm:"column:password"`
	CreateDate time.Time `json:"createDate" gorm:"column:createdate"`
	ModifyDate time.Time `json:"modifyDate" gorm:"column:modifydate"`
}

func (User) TableName() string {
	return "tbl_user"
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Fullname string `json:"fullname" validate:"required"`
	UserType string `json:"usertype" validate:"required"`
	Password string `json:"password" validate:"required"`
}
