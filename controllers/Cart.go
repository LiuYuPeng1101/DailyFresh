package controllers

import (
	"ShFresh/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"strconv"
)

type Cartcontrollers struct {
	beego.Controller
}
//购物车图标
func (this*Cartcontrollers)HandleAddCart(){
	SKUId,err1:=this.GetInt("SKUId")
	Count,err2:=this.GetInt("Count")
	resp:=make(map[string]interface{})
	defer this.ServeJSON()
	if err1 !=nil || err2!=nil{
		resp["code"] = 1
		resp["msg"] = "获取数据失败"
		this.Data["json"] =resp
		beego.Info("数据传递失败")
	}
	UserName:=this.GetSession("username")
	if UserName == nil{
		resp["code"] = 2
		resp["msg"] = "用户没登录"
		this.Data["json"] =resp
		beego.Info("用户没登录")
	}
	o:=orm.NewOrm()
	var User models.User
	User.Name = UserName.(string)
	o.Read(&User,"Name")
	//连接redis
	Conn,err:=redis.Dial("tcp","172.16.107.175:6379")
	if err !=nil{
		beego.Info("连接数据库失败")
	}
	defer Conn.Close()
	//获取商品数量
	PreCount,_:=redis.Int(Conn.Do("hget","Cart"+strconv.Itoa(User.Id),SKUId))
	//把用户信息、商品信息、数量添加进redis
	Conn.Do("hset","Cart_"+strconv.Itoa(User.Id),SKUId,Count+PreCount)
	res,err:=Conn.Do("hlen","Cart_"+strconv.Itoa(User.Id))
	CartCount,err:=redis.Int(res,err)
	//beego中怎么给视图返回json数据,可以使用Map
	resp["code"] = 5
	resp["msg"] = "ok"
	resp["CartCount"] = CartCount
	this.Data["json"] =resp
}
//封装购物车数量函数
func GetCartCount(this*beego.Controller)int{
	UserName:=this.GetSession("username")
	if UserName == nil{
		return 0
	}
	o:=orm.NewOrm()
	var User models.User
	User.Name = UserName.(string)
	o.Read(&User,"Name")
	Conn,err:=redis.Dial("tcp","172.16.107.175:6379")
	if err !=nil{
		beego.Info("连接数据库失败")
	}
	res,err:=Conn.Do("hlen","Cart_"+strconv.Itoa(User.Id))
	CartCount,_:=redis.Int(res,err)
	return CartCount
}
//购物车
func (this*Cartcontrollers)ShowCart(){
	UserName:=this.GetSession("username")
	o:=orm.NewOrm()
	var User models.User
	User.Name = UserName.(string)
	o.Read(&User,"Name")
	//连接redis
	Conn,err:=redis.Dial("tcp","172.16.107.175:6379")
	if err !=nil{
		beego.Info("连接数据库失败")
	}
	defer Conn.Close()
	GoodsMap,err:=redis.IntMap(Conn.Do("hgetall","Cart_"+strconv.Itoa(User.Id)))
	Goods:=make([]map[string]interface{},len(GoodsMap))
	i:=0
	TotalPrice :=0
	TotalCount :=0
	for index,value:=range GoodsMap{
		SKUId,_:=strconv.Atoi(index)
		var GoodsSKU models.GoodsSKU
		GoodsSKU.Id = SKUId
		o.Read(&GoodsSKU)
		temp:=make(map[string]interface{})
		temp["Goods"] = GoodsSKU
		temp["Count"] = value
		TotalPrice += GoodsSKU.Price * value
		TotalCount += value
		temp["addPrice"] = GoodsSKU.Price * value
		Goods[i] = temp
		i+=1
	}
	//返回视图
	this.Data["TotalPrice"] = TotalPrice
	this.Data["TotalCount"] = TotalCount
	this.Data["Goods"] = Goods
	this.TplName = "cart.html"
}
//更新商品数量
func (this*Cartcontrollers)HandleUpdateCart(){
	SKUId,err1:=this.GetInt("skuid")
	Count,err2:=this.GetInt("count")
	resp:=make(map[string]interface{})
	defer this.ServeJSON()
	if err1!=nil || err2!=nil{
		resp["code"] = 1
		resp["msg"] = "获取数据失败"
		this.Data["json"] = resp
		beego.Info("获取数据失败")
		return
	}
	UserName:=this.GetSession("username")
	if UserName == nil{
		resp["code"] = 2
		resp["msg"] = "用户没登录"
		this.Data["json"] = resp
		beego.Info("用户没登录")
		return
	}
	o:=orm.NewOrm()
	var User models.User
	User.Name = UserName.(string)
	o.Read(&User,"Name")
	//连接redis
	Conn,err:=redis.Dial("tcp","172.16.107.175:6379")
	if err !=nil{
		beego.Info("连接数据库失败")
	}
	defer Conn.Close()

	Conn.Do("hset","Cart_"+strconv.Itoa(User.Id),SKUId,Count)
	resp["code"] =5
	resp["msg"] = "ok"
	this.Data["json"] = resp
}
//删除商品
func (this*Cartcontrollers)HandleDeleteCart(){
	SKUId,err:=this.GetInt("skuid")
	resp:=make(map[string]interface{})
	defer this.ServeJSON()
	if err!=nil{
		resp["code"] = 1
		resp["msg"] = "获取数据失败"
		this.Data["json"] = resp
		beego.Info("获取数据失败")
		return
	}
	UserName:=this.GetSession("username")
	if UserName == nil{
		resp["code"] = 2
		resp["msg"] = "用户没登录"
		this.Data["json"] = resp
		beego.Info("用户没登录")
		return
	}
	o:=orm.NewOrm()
	var User models.User
	User.Name = UserName.(string)
	o.Read(&User,"Name")
	//连接redis
	Conn,err:=redis.Dial("tcp","172.16.107.175:6379")
	if err !=nil{
		beego.Info("连接数据库失败")
	}
	defer Conn.Close()

	Conn.Do("hdel","Cart_"+strconv.Itoa(User.Id),SKUId)
	resp["code"] =5
	resp["msg"] = "ok"
	this.Data["json"] = resp
}
