package models

import (
	//"fmt"
	"time"
)

// User definition
type User struct {
	Id                  int                   `orm:"auto;index"`
	Email               string                `orm:"size(64);index;unique"`
	Password            string                `orm:"size(128)"`
	Name                string                `orm:"size(64);index"`
	Nickname            string                `orm:"size(64);unique;index"`
	LastLogin           time.Time             `orm:"auto_now_add"`
	DateJoined          time.Time             `orm:"auto_now_add"`
	IsActive            bool                  `orm:"default(1)"`
	IsAdmin             bool                  `orm:"default(0);index"`
	Profile             *UserProfile          `orm:"reverse(one)"`
	Permissions         []*Permission         `orm:"rel(m2m);rel_table(user_permission)"`
	PasswordPermissions []*PasswordPermission `orm:"reverse(many)"`
	//Passwords []*PasswordInfo `orm:"rel(m2m)"`
}

// User Additional info
type UserProfile struct {
	Id         int    `orm:"auto;index"`
	Department string `orm:"index"`
	Title      string `orm:"null"`
	Mobile     string `orm:"index"`
	Phone      string `orm:"null;index"`
	User       *User  `orm:"rel(one)"`
	Salt       string
}

// Register invitation
type RegisterInvitation struct {
	Id        int `orm:"auto"`
	Token     string
	Email     string
	Expired   bool
	IssueDate time.Time `orm:"auto_now_add"`
}

type Permission struct {
	Id            int `orm:"auto"`
	ContentTypeId int
	Name          string
	Codename      string
	Users         []*User `orm:"reverse(many)"`
}

type PasswordInfo struct {
	Id          int                   `orm:"auto;index"`
	Name        string                `orm:"index" form:"name" valid:"Required"`
	Account     string                `form:"account" valid:"Required"`
	Password    string                `form:"password" valid:"Required"`
	Desc        string                `orm:"null" form:"desc"`
	Permissions []*PasswordPermission `orm:"reverse(many)"`
	//Users []*User `orm:"reverse(many)"`
}

const (
	NoPermission = iota
	CanRead
	CanUpdate
	CanManage
)

type PasswordPermission struct {
	Id       int           `orm:"auto;index`
	Password *PasswordInfo `orm:"rel(fk)"`
	User     *User         `orm:"rel(fk)"`
	Level    int
}

func (this *PasswordPermission) CanRead() bool {
	return this.Level >= CanRead
}

func (this *PasswordPermission) CanUpdate() bool {
	return this.Level >= CanUpdate
}

func (this *PasswordPermission) CanManage() bool {
	return this.Level >= CanManage
}

func (this *PasswordPermission) TableUnique() [][]string {
	return [][]string{
		[]string{"Password", "User"},
	}
}

type MPKey struct {
	Id      int `orm:"auto"`
	DataKey string
}

type MPApiToken struct {
	Id    int `orm:"auto"`
	Token string
}

type Tab struct {
	TabName string
}

func (this *Tab) IsIndex() bool {
	return this.TabName == "Index"
}

func (this *Tab) IsPassword() bool {
	return this.TabName == "Password"
}

func (this *Tab) IsScheduler() bool {
	return this.TabName == "Scheduler"
}

func (this *Tab) IsBrand() bool {
	return this.TabName == "brand"
}
func (this *Tab) IsBlack() bool {
	return this.TabName == "blacklist"
}
func (this *Tab) IsCommodity() bool {
	return this.TabName == "Commodity"
}

func (this *Tab) IsProfile() bool {
	return this.TabName == "Profile"
}

func (this *Tab) IsWordManager() bool {
	return this.TabName == "Words"
}
