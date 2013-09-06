package controllers
import (
	//"fmt"
	"strconv"
	//"time"

	"Mango/management/models"
	//"Mango/management/utils"

	//"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	//"github.com/astaxie/beego/validation"
	//_ "github.com/go-sql-driver/mysql"
)


type ListPassController struct {
    UserSessionController
}


func (this *ListPassController) Get() {
    user := this.Data["User"].(*models.User)
    p := models.PasswordPermission{}
    o := orm.NewOrm()
    //var passwords []int
    var passwords []*models.PasswordPermission
    o.QueryTable(&p).Filter("User", user.Id).All(&passwords)
    result := ""
    for _, v := range passwords {
        result += strconv.Itoa(v.Password.Id) + " "
    }
    this.Ctx.WriteString(result)
}


type AddPassController struct {
    UserSessionController
}

func (this *AddPassController) Get() {
}

func (this *AddPassController) Post() {
}

type EditPassController struct {
    UserSessionController
}

func (this *EditPassController) Get() {
}

func (this *EditPassController) Post() {
}

type DeletePassController struct {
    UserSessionController
}

func (this *DeletePassController) Get() {
}

func (this *DeletePassController) Post() {
}




