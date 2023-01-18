package model

type User struct {
	Username string `json:"username" binding:"required" bson:"username"`
	Password string `json:"-" binding:"required" bson:"password"`
	Saying   string `json:"saying" bson:"saying"`
}

type UserLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserByUserName struct {
	Username string `json:"username" binding:"required"`
}