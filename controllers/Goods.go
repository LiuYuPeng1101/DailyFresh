package controllers

import (
	"ShFresh/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"math"
	"strconv"
)

type Goodscontrollers struct {
	beego.Controller
}
//封装GetSession函数,因为每一个登录的页面都要用到
func GetUser(this*beego.Controller)string{
	UserName:=this.GetSession("username")
	if UserName == nil{
		this.Data["username"] = ""
		return ""
	}else{
		this.Data["username"] = UserName.(string)
		return UserName.(string)
	}
}
//展示首页
func (this*Goodscontrollers)ShowIndex(){
	GetUser(&this.Controller)
	o:=orm.NewOrm()
	//展示首页商品分类
	var GoodsType []models.GoodsType
	o.QueryTable("GoodsType").All(&GoodsType)
	this.Data["GoodsType"] = GoodsType
	//展示首页伦播图片
	var IndexGoodsBanner []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&IndexGoodsBanner)
	this.Data["IndexGoodsBanner"] = IndexGoodsBanner
	//展示首页促销图片
	var IndexPromotionBanner []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&IndexPromotionBanner)
	this.Data["IndexPromotionBanner"] = IndexPromotionBanner
	//按分类展示首页商品数据
	Goods:=make([]map[string]interface{},len(GoodsType))
	for index,value := range GoodsType{
		temp := make(map[string]interface{})
		temp["types"] = value
		Goods[index] =temp
	}
	for _,value:= range Goods{
		var ImgGoods []models.IndexTypeGoodsBanner
		var TextGoods []models.IndexTypeGoodsBanner
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSkU").Filter("GoodsType",value["types"]).Filter("DisplayType",1).OrderBy("Index").All(&ImgGoods)
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSkU").Filter("GoodsType",value["types"]).Filter("DisplayType",0).OrderBy("Index").All(&TextGoods)
		value["ImgGoods"] = ImgGoods
		value["TextGoods"] = TextGoods
	}
	this.Data["Goods"] = Goods
	//返回视图
	this.TplName = "index.html"
}
//封装函数:商品类型展示、session
func ShowLayout(this*beego.Controller){
	o:=orm.NewOrm()
	var GoodsType []models.GoodsType
	o.QueryTable("GoodsType").All(&GoodsType)
	this.Data["types"] = GoodsType
	GetUser(this)
	this.Layout="GoodsDetailLayout.html"
}
//商品详情页
func (this*Goodscontrollers)ShowGoodsDetail(){
	ShowLayout(&this.Controller)
	//获取数据
	Id,err:=this.GetInt("Id")
	if err !=nil{
		beego.Error("获取数据失败")
		this.Redirect("/",302)
		return
	}
	//操作数据
	o:=orm.NewOrm()
	var GoodsSKU models.GoodsSKU
	GoodsSKU.Id = Id
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType","Goods").Filter("Id",Id).One(&GoodsSKU)


	//新品推荐
	var NewGoods []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name",&GoodsSKU.GoodsType.Name).Limit(2).OrderBy("time").All(&NewGoods)

	//最近浏览
	UserName:=GetUser(&this.Controller)
	if UserName!=""{
		o:=orm.NewOrm()
		var User models.User
		User.Name = UserName
		o.Read(&User,"Name")
		//连接redis
		Conn,err:=redis.Dial("tcp","172.16.107.175:6379")
		if err !=nil{
			beego.Error("连接数据库失败")
		}
		defer Conn.Close()
		Conn.Do("lrem","history_"+strconv.Itoa(User.Id),Id,Id)
		Conn.Do("lpush","history_"+strconv.Itoa(User.Id),Id)
	}
	//购物车数量
	CartCount:=GetCartCount(&this.Controller)
	//返回视图
	this.Data["CartCount"] = CartCount
	this.Data["GoodsSKU"] = GoodsSKU
	this.Data["NewGoods"] = NewGoods
	this.TplName = "detail.html"
}
//封装分页函数
func PageTool(PageCount int,PageIndex int)[]int{
	var Pages = []int{}
	if PageCount < 5{
		Pages=make([]int,PageCount)
		for i,_:=range Pages{
			Pages[i] = i+1
		}
	}else if PageIndex <= 3{
		Pages=[]int{1,2,3,4,5}
	}else if PageIndex > PageCount - 3{
		Pages=[]int{PageCount-4,PageCount-3,PageCount-2,PageCount-1,PageCount}
	}else{
		Pages=[]int{PageCount-2,PageCount-1,PageCount,PageCount+1,PageCount+2}
	}
	return Pages
}
//商品列表页
func (this*Goodscontrollers)ShowGoodsList(){
	ShowLayout(&this.Controller)
	Id,err:=this.GetInt("Id")
	if err !=nil{
		beego.Info("获取ID失败",err)
		beego.Info(Id)
		this.Redirect("/",302)
		return
	}
	o:=orm.NewOrm()
	//分页处理
	Count,_:=o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",&Id).Count()
	PageSize:=2
	PageCount:=math.Ceil(float64(Count)/float64(PageSize))
	PageIndex,err:=this.GetInt("PageIndex")
	if err !=nil{
		PageIndex = 1
	}
	Pages:=PageTool(int(PageCount),PageIndex)
	//上一页和下一页
	PrePage:=PageIndex -1
	if PrePage <= 0{
		PrePage = 1
	}
	NextPage:=PageIndex +1
	if NextPage >int(PageCount){
		NextPage = int(PageCount)
	}
	//默认排序、价格排序、人气排序
	start:=(PageIndex-1)*PageSize
	var DefaultGoods []models.GoodsSKU
	sort:=this.GetString("sort")
	if sort == ""{
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",&Id).Limit(PageSize,start).All(&DefaultGoods)
		this.Data["sort"] = ""
		this.Data["DefaultGoods"] = DefaultGoods
	}else if sort == "Price"{
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__id",&Id).OrderBy("Price").Limit(PageSize,start).All(&DefaultGoods)
		this.Data["sort"] = "Price"
		this.Data["DefaultGoods"] = DefaultGoods
	}else{
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__id",&Id).OrderBy("sales").Limit(PageSize,start).All(&DefaultGoods)
		this.Data["sort"] = "Sales"
		this.Data["DefaultGoods"] = DefaultGoods
	}
	//新品推荐
	var NewGoods []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",&Id).OrderBy("time").Limit(2).All(&NewGoods)
	//返回视图
	this.Data["PrePage"] = PrePage
	this.Data["NextPage"] = NextPage
	this.Data["PageIndex"] = PageIndex
	this.Data["Id"] = Id
	this.Data["Pages"] = Pages
	this.Data["NewGoods"] = NewGoods
	this.Layout = "GoodsDetailLayout.html"
	this.TplName = "list.html"
}
//搜索
func (this*Goodscontrollers)HandleGoodsSearch(){
	GoodsName:=this.GetString("GoodsName")
	o:=orm.NewOrm()
	var GoodsSKU []models.GoodsSKU
	if GoodsName == ""{
		o.QueryTable("GoodsSKU").All(&GoodsSKU)
		this.Data["GoodsSKU"] = GoodsSKU
		ShowLayout(&this.Controller)
		this.TplName = "Search.html"
		return
	}
	o.QueryTable("GoodsSKU").Filter("Name__icontains",GoodsName).All(&GoodsSKU)
	this.Data["GoodsSKU"] = GoodsSKU
	ShowLayout(&this.Controller)
	this.TplName = "Search.html"
}
