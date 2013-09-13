package controllers

import (
    "Mango/management/models"
    "github.com/astaxie/beego/orm"
)

func CheckPermission(user_id int, permissionCode string) bool {
    o := orm.NewOrm()
    perm := models.Permission{}
    o.QueryTable(&perm).Filter("Users__Id", user_id).Filter("Codename", permissionCode).One(&perm)
    if perm.Id != 0 {
        return true
    }
    return false
}
