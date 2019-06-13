package controllers

import (
	"ShFresh/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"strings"
	"time"
)

type OrderControllers struct {
	beego.Controller
}

//展示订单页面
func (this*OrderControllers)ShowOrder(){
	SKUIds:=this.GetStrings("skuid")
	Goods:=make([]map[string]interface{},len(SKUIds))
	if len(SKUIds) == 0{
		this.Redirect("/User/Cart",302)
		beego.Info("获取数据失败")
		return
	}
	UserName:=this.GetSession("username")
	if UserName == nil{
		this.Redirect("/Login",302)
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
	i:=0
	TotalCount:= 0
	TotalPrice:=0
	for _,value :=range SKUIds{
		resp:=make(map[string]interface{})
		o:=orm.NewOrm()
		var Address  []models.Address
		o.QueryTable("Address").RelatedSel("User").Filter("User__Name",&User.Name).All(&Address)
		this.Data["Address"] = Address
		var GoodsSKU models.GoodsSKU
		o.QueryTable("GoodsSKU").Filter("Id",value).One(&GoodsSKU)
		resp["Goods"] = GoodsSKU
		Count,_:=redis.Int(Conn.Do("hget","Cart_"+strconv.Itoa(User.Id),value))
		resp["Count"] = Count
		//小计
		GoodsPrice:=float64(GoodsSKU.Price * Count)
		resp["GoodsPrice"] = GoodsPrice
		//商品id
		resp["Value"] = value
		Goods[i]=resp
		i+=1
		TotalCount+=Count
		TotalPrice+=GoodsSKU.Price
	}
	//实付款
	freight:=10
	TotalPay:=TotalPrice+freight
	//返回视图
	this.Data["SKUId"] = SKUIds
	this.Data["freight"] = freight
	this.Data["TotalPay"] = TotalPay
	this.Data["TotalCount"] = TotalCount
	this.Data["TotalPrice"] = float64(TotalPrice)
	this.Data["Goods"] = Goods
	this.TplName = "place_order.html"


}
//添加订单
func (this*OrderControllers)AddOrder(){
	//获取数据
	addrid, _ := this.GetInt("addrid")
	payId, _ := this.GetInt("payId")
	skuid := this.GetString("skuids")
	ids:=skuid[1:len(skuid)-1]
	skuids:=strings.Split(ids,"")
	//totalPrice,_ := this.GetInt("totalPrice")
	totalCount, _ := this.GetInt("totalCount")
	transferPrice, _ := this.GetInt("transferPrice")
	realyPrice, _ := this.GetInt("realyPrice")

	resp := make(map[string]interface{})
	defer this.ServeJSON()
	//校验数据
	if len(skuids) == 0 {
		resp["code"] = 1
		resp["errmsg"] = "数据库链接错误"
		this.Data["json"] = resp
		return
	}
	//处理数据
	//向订单表中插入数据
	o := orm.NewOrm()
	o.Begin()
	userName := this.GetSession("username")
	var user models.User
	user.Name = userName.(string)
	o.Read(&user,"Name")

	var order models.OrderInfo
	order.OrderId = time.Now().Format("2006010215030405")+strconv.Itoa(user.Id)
	order.User = &user
	order.Orderstatus = 1
	order.PayMethod = payId
	order.TotalCount = totalCount
	order.TotalPrice = realyPrice
	order.TransitPrice = transferPrice
	//查询地址
	var addr models.Address
	addr.Id = addrid
	o.Read(&addr)

	order.Address = &addr

	//执行插入操作
	o.Insert(&order)
	//想订单商品表中插入数据
	conn,_ :=redis.Dial("tcp","172.16.107.175:6379")

	for _,skuid := range skuids {
		id, _ := strconv.Atoi(skuid)

		var goods models.GoodsSKU
		goods.Id = id
		i:=3
		if i>0 {
			o.Read(&goods)
			var orderGoods models.OrderGoods
			orderGoods.GoodsSKU = &goods
			orderGoods.OrderInfo = &order
			count, _ := redis.Int(conn.Do("hget", "Cart_"+strconv.Itoa(user.Id), id))
			if count > goods.Stock {
				resp["code"] = 2
				resp["errmsg"] = "商品库存不足"
				this.Data["json"] = resp
				o.Rollback()
				return
			}
			precount := goods.Stock
			orderGoods.Count = count
			orderGoods.Price = count * goods.Price
			o.Insert(&orderGoods)
			goods.Stock -= count
			goods.Sales += count
			UpdateCount, _ := o.QueryTable("GoodsSKU").Filter("Id", goods.Id).Filter("stock", precount).Update(orm.Params{"stock": goods.Stock, "sales": goods.Sales})
			if UpdateCount == 0 {
				if i >0 {
					i -= 1
					continue
				}
				resp["code"] = 3
				resp["errmsg"] = "商品库存不足"
				this.Data["json"] = resp
				o.Rollback()
				return
			}else{
				conn.Do("hdel","cart_"+strconv.Itoa(user.Id),goods.Id)
				break
			}
		}
	}
	o.Commit()
	resp["code"] = 5
	resp["errmsg"] = "ok"
	this.Data["json"] = resp

}
