package controllers

import (
	"ShFresh/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils"
	"github.com/gomodule/redigo/redis"
	"regexp"
	"strconv"
)

type Usercontrollers struct {
	beego.Controller
}

//展示注册页面
func (this*Usercontrollers)ShowRegister(){
	//给前端传递视图
	this.TplName = "register.html"
}
//处理注册页面
func (this*Usercontrollers)HandleRegister(){
	//获取数据
	UserName:=this.GetString("user_name")
	Pwd:=this.GetString("pwd")
	Cpwd:=this.GetString("cpwd")
	Email:=this.GetString("email")
	//校验数据
	if UserName == "" || Pwd == "" || Cpwd == "" || Email == ""{
		this.Data["errmsg"] = "您输入的数据不完整,请重新输入"
		this.TplName = "register.html"
		return
	}
	if Pwd != Cpwd{
		this.Data["errmsg"] = "您两次输入的密码不一致,请重新输入"
		this.TplName = "register.html"
		return
	}
	reg,_:=regexp.Compile("^[A-Za-z0-9\u4e00-\u9fa5]+@[a-zA-Z0-9_-]+(\\.[a-zA-Z0-9_-]+)+$")
	res:=reg.FindString(Email)
	if res == ""{
		this.Data["errmsg"] = "您输入的邮箱格式不正确,请重新输入"
		this.TplName = "register.html"
		return
	}
	//操作数据
	o:=orm.NewOrm()
	var User models.User
	User.Name = UserName
	User.PassWord = Pwd
	User.Email = Email

	_,err:=o.Insert(&User)
	if err !=nil{
		this.Data["errmsg"] = "注册失败,请更换您的数据"
		this.TplName = "register.html"
		return
	}
	//发送邮箱验证连接
	EmailConf:=`{"username":"729871301@qq.com","password":"fdouuxkbqxvrbdce","host":"smtp.qq.com","port":587}`
	EmailConn:=utils.NewEMail(EmailConf)
	beego.Info(EmailConn)
	EmailConn.From = "729871301@qq.com"
	EmailConn.To = []string{Email}
	EmailConn.Subject = "天天生鲜用户注册"
	EmailConn.Text = "尊敬的" +User.Name +":"+"\n"+
		"    "+"感谢您的注册,请点击下方链接进行激活."+"\n"+"    "+
		"http://172.16.107.175:8080/Active?Id="+strconv.Itoa(User.Id)
	err=EmailConn.Send()
	if err !=nil{
		this.Data["errmsg"] = "发送邮件失败,请重新注册"
		this.TplName = "register.html"
		return
	}
	//返回视图
	this.Ctx.WriteString("注册成功,请前往您的邮箱激活！")
}
//处理激活页面
func (this*Usercontrollers)ActiveUser(){
	//获取数据
	Id,err:=this.GetInt("Id")
	if err !=nil{
		this.Data["errmsg"] = "您要激活的账户不存在,请重新激活"
		this.TplName = "register.html"
		return
	}
	//操作数据
	o:=orm.NewOrm()
	var User models.User
	User.Id = Id
	err=o.Read(&User)
	if err !=nil{
		this.Data["errmsg"] = "您要激活的账户不存在,请重新激活"
		this.TplName = "register.html"
		return
	}
	User.Active = true
	_,err=o.Update(&User)
	if err !=nil{
		this.Data["errmsg"] = "您的账户激活失败,请重新激活"
		this.TplName = "register.html"
		return
	}
	//返回视图
	this.Redirect("/Login",302)
}
//展示登录页面
func (this*Usercontrollers)ShowLogin(){
	UserName:=this.Ctx.GetCookie("username")
	if UserName != ""{
		this.Data["UserName"] = UserName
		this.Data["checked"] = "checked"
	}else{
		this.Data["UserName"] = ""
		this.Data["checked"] = ""
	}
	this.TplName = "login.html"
}
//处理登录页面
func (this*Usercontrollers)HandleLogin(){
	//获取数据
	UserName:=this.GetString("username")
	Pwd:=this.GetString("pwd")
	//校验数据
	if UserName == "" || Pwd == ""{
		this.Data["errmsg"] = "请输入用户名和密码"
		this.TplName = "login.html"
		return
	}
	o:=orm.NewOrm()
	var User models.User
	User.Name = UserName
	err:=o.Read(&User,"Name")
	if err !=nil{
		this.Data["errmsg"] = "用户名不存在,请重新输入"
		this.TplName = "login.html"
		return
	}
	if User.PassWord !=Pwd{
		this.Data["errmsg"] = "您输入的密码不正确,请重新输入"
		this.TplName = "login.html"
		return
	}
	if User.Active != true{
		this.Data["errmsg"] = "您输入的用户账号未激活,请激活登录"
		this.TplName = "login.html"
		return
	}
	//操作数据:设置cookie
	Remember:=this.GetString("remember")
	if Remember == "on"{
		this.Ctx.SetCookie("username",UserName,365*24*60*60)
	}else{
		this.Ctx.SetCookie("username",UserName,-1)
	}
	//设置session
	this.SetSession("username",UserName)
	//返回视图
	this.Redirect("/",302)
}
//退出处理
func (this*Usercontrollers)ShowLogout(){
	this.DelSession("username")
	this.Redirect("/Login",302)
}
//用户中心--个人信息
func (this*Usercontrollers)ShowUserCenterInfo(){
	UserName:=GetUser(&this.Controller)
	this.Data["username"] = UserName
	//操作数据库
	o:=orm.NewOrm()
	var Addr models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name",UserName).Filter("IsDefault",true).One(&Addr)
	if Addr.Id == 0 {
		this.Data["Addr"] = ""
	}else {
		this.Data["Addr"] = Addr
	}
	//最近浏览
	Conn,err:=redis.Dial("tcp","172.16.107.175:6379")
	if err !=nil{
		beego.Error("连接数据库失败")
	}
	defer Conn.Close()

	var User models.User
	User.Name = UserName
	beego.Info(UserName)
	o.Read(&User,"Name")
	rep,err:=Conn.Do("lrange","history_"+strconv.Itoa(User.Id),0,4)
	GoodsId,err:=redis.Ints(rep,err)
	var GoodsSKUs []models.GoodsSKU
	for _,value:=range GoodsId{
		var GoodsSKU models.GoodsSKU
		GoodsSKU.Id = value
		o.Read(&GoodsSKU)
		beego.Info(GoodsSKU)
		GoodsSKUs=append(GoodsSKUs,GoodsSKU)
	}
	beego.Info(GoodsSKUs)
	this.Data["GoodsSKUs"] = GoodsSKUs
	this.Layout = "UserCenterLayout.html"
	this.TplName = "user_center_info.html"
}
//用户中心--全部订单
func(this*Usercontrollers)ShowUserCenterOrder(){
	UserName := GetUser(&this.Controller)
	o := orm.NewOrm()
	var user models.User
	user.Name = UserName
	o.Read(&user,"Name")
	//获取订单表的数据
	var orderInfos []models.OrderInfo
	o.QueryTable("OrderInfo").RelatedSel("User").Filter("User__Id",user.Id).All(&orderInfos)
	goodsBuffer := make([]map[string]interface{},len(orderInfos))
	for index,orderInfo := range orderInfos{
		var orderGoods []models.OrderGoods
		o.QueryTable("OrderGoods").RelatedSel("OrderInfo","GoodsSKU").Filter("OrderInfo__Id",orderInfo.Id).All(&orderGoods)
		temp := make(map[string]interface{})
		temp["orderInfo"] = orderInfo
		temp["orderGoods"] = orderGoods
		goodsBuffer[index] = temp
	}
	this.Data["goodsBuffer"] = goodsBuffer
	//订单商品表
	this.Layout = "UserCenterLayout.html"
	this.TplName = "user_center_order.html"
}
//用户中心--收货地址
func (this*Usercontrollers)ShowUserCenterSite(){
	UserName:=GetUser(&this.Controller)
	//操作数据库
	o:=orm.NewOrm()
	var Address models.Address
	qs:=o.QueryTable("Address").RelatedSel("User").Filter("User__Name",UserName).Filter("IsDefault",true).One(&Address)
	beego.Info(qs)
	this.Data["Address"] = Address
	this.Layout = "UserCenterLayout.html"
	this.TplName = "user_center_site.html"
}
//用户中心--收货地址
func (this*Usercontrollers)HandleUserCenterSite(){
	//获取数据
	Receiver:=this.GetString("Receiver")
	Addr:=this.GetString("Addr")
	ZipCode:=this.GetString("ZipCode")
	Phone:=this.GetString("Phone")
	//校验数据
	if Receiver == "" || Addr == "" || ZipCode == "" || Phone == ""{
		beego.Error("输入的数据为空")
		this.Redirect("/User/user_center_site.html",302)
		return
	}
	//操作数据
	o:=orm.NewOrm()
	var Address models.Address
	Address.Isdefault = true
	err:=o.Read(&Address,"IsDefault")
	if err ==nil {
		beego.Info("ppppp")
		Address.Isdefault = false
		o.Update(&Address)
	}
	UserName:=GetUser(&this.Controller)
	var User models.User
	User.Name = UserName
	o.Read(&User,"Name")
	var AddressNew models.Address
	AddressNew.Receiver = Receiver
	AddressNew.Addr = Addr
	AddressNew.Zipcode = ZipCode
	AddressNew.Phone = Phone
	AddressNew.Isdefault= true
	AddressNew.User = &User
	o.Insert(&AddressNew)
	//返回视图
	this.Redirect("/User/user_center_site.html",302)
}
