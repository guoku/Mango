package controllers

import (
	"Mango/management/models"
	"github.com/astaxie/beego/orm"
)

func CheckPermission(user_id int, permissionCode string) bool {
	o := orm.NewOrm()
	/*perm := models.Permission{}
	o.QueryTable(&perm).Filter("Users__Id", user_id).Filter("Codename", permissionCode).One(&perm)
	if perm.Id != 0 {
		return true
	}*/
    permission := models.Permission{Codename : permissionCode}
    err := o.Read(&permission, "Codename")
    if err != nil {
        return false
    }
    user := models.User{Id : user_id}
    m2m := o.QueryM2M(&user, "Permissions")
    if m2m.Exist(&permission) {
        return true
    }
	return false
}
