package model

type User struct {
	Uid      int64  `json:"uid" bson:"uid"`
	Username string `json:"username" bson:"username"`
	Password string `json:"-" bson:"password"`
	Email    string `json:"email" bson:"email"`
	Phone    string `json:"phone" bson:"phone"`
	Avatar   string `json:"avatar" bson:"avatar"`
	Role     int64  `json:"-" bson:"role"`
	Disabled bool   `json:"-" bson:"disabled" default:"false"`
}

type UserLogin struct {
	Uid      int64  `json:"uid" bson:"uid" binding:"required"`
	Password string `json:"password" binding:"required" bson:"password"`
}

type UserRegister struct {
	Email    string `json:"email" binding:"required" bson:"email"`
	Password string `json:"password" binding:"required" bson:"password"`
}

type UserByUid struct {
	Uid int64 `json:"uid" bson:"uid" binding:"required"`
}
