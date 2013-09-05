package models

import "time"

// User definition
type User struct {
	Id         int       `orm:"auto"`
	Email      string    `orm:"size(64)"`
	Password   string    `orm:"size(128)"`
	Name       string    `orm:"size(64)"`
	Nickname   string    `orm:"size(64)"`
	LastLogin  time.Time `orm:"auto_now_add"`
	DateJoined time.Time `orm:"auto_now_add"`
	IsActive   bool      `orm:"default(true)"`
	IsAdmin    bool
    Additional  *UserAdditional `orm:"reverse(one)"` 
}

func (this *User) TableName() string {
	return "staff_user"
}

func (this *User) TableIndex() [][]string {
	return [][]string{
		[]string{"Email", "Nickname"},
	}
}

func (this *User) TableUnique() [][]string {
	return [][]string{
		[]string{"Nickname", "Email"},
	}
}

// User Additional info
type UserAdditional struct {
	Id         int    `orm:"auto"`
	Department string `orm:"null"`
	Title      string `orm:"null"`
	Mobile     string `orm:"null"`
	Phone      string `orm:"null"`
	User       *User  `orm:"rel(one)"`
	Salt       string
}

func (this *UserAdditional) TableName() string {
	return "user_additional_info"
}

// Register invitation
type RegisterInvitation struct {
	Id        int `orm:"auto"`
	Token     string
	Email     string
	Expired   bool
	IssueDate time.Time     `orm:"auto_now_add"`
}

func (this *RegisterInvitation) TableName() string {
	return "register_invitation"
}
