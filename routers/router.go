package routers

import (
	"ShFresh/controllers"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego"
)

func init() {
	beego.InsertFilter("/User/*",beego.BeforeExec,FilterFunc)
    //beego.Router("/", &controllers.MainController{})
    //注册
    beego.Router("/Register",&controllers.Usercontrollers{},"get:ShowRegister;post:HandleRegister")
    //激活
    beego.Router("/Active",&controllers.Usercontrollers{},"get:ActiveUser")
    //登录
    beego.Router("/Login",&controllers.Usercontrollers{},"get:ShowLogin;post:HandleLogin")
    //首页
    beego.Router("/",&controllers.Goodscontrollers{},"get:ShowIndex")
	//退出
	beego.Router("/User/Logout",&controllers.Usercontrollers{},"get:ShowLogout")
	//用户中心--个人信息
	beego.Router("/User/user_center_info.html",&controllers.Usercontrollers{},"get:ShowUserCenterInfo")
	//用户中心--全部订单
	beego.Router("/User/user_center_order.html",&controllers.Usercontrollers{},"get:ShowUserCenterOrder")
	//用户中心--收货地址
	beego.Router("/User/user_center_site.html",&controllers.Usercontrollers{},"get:ShowUserCenterSite;post:HandleUserCenterSite")
	//商品详情页
	beego.Router("/GoodsDetail",&controllers.Goodscontrollers{},"get:ShowGoodsDetail")
	//商品列表页面
	beego.Router("/GoodsList",&controllers.Goodscontrollers{},"get:ShowGoodsList")
	//搜索页面
	beego.Router("/GoodsSearch",&controllers.Goodscontrollers{},"post:HandleGoodsSearch")
	//购物车图标
	beego.Router("/User/AddCart",&controllers.Cartcontrollers{},"post:HandleAddCart")
	//购物车
	beego.Router("/User/Cart",&controllers.Cartcontrollers{},"get:ShowCart")
	//更新商品数量
	beego.Router("/User/UpdateCart",&controllers.Cartcontrollers{},"post:HandleUpdateCart")
	//删除商品
	beego.Router("/User/DeleteCart",&controllers.Cartcontrollers{},"post:HandleDeleteCart")
	//展示订单页面
	beego.Router("/User/ShowOrder",&controllers.OrderControllers{},"post:ShowOrder")
	//添加订单
	beego.Router("/User/AddOrder",&controllers.OrderControllers{},"post:AddOrder")
	//支付
}

var FilterFunc = func(Ctx*context.Context) {
	UserName:=Ctx.Input.Session("username")
	if UserName == nil{
		Ctx.Redirect(302,"/Login")
		return
	}
}
