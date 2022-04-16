package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)


type AuthParam struct {
	Name string `json:"name" binding:"required"`	// binding:"required" 代表前端必须要传这个字段
	Password string `json:"password" binding:"required"`
}


// loginHandler 用户登录的函数
func loginHandler(c *gin.Context){
	// 1，获取参数，校验参数
	var param AuthParam
	if err := c.ShouldBind(&param);err != nil{
		c.JSON(http.StatusOK, Resp{
			Code: 1,
			Msg: "参数错误",
		})
		return
	}

	// 2，逻辑处理
	// 拿 用户名和密码，去数据库校验，能查到记录代表登录成功
	var u Account
	// 将查到的数据，赋值到u结构体中
	err := db.Where("name=? and password=?",param.Name,md5secret(param.Password)).First(&u).Error
	if err != nil{
		if errors.Is(err, gorm.ErrRecordNotFound){
			c.JSON(http.StatusOK, Resp{
				Code: 1,
				Msg: "用户名或者密码错误",
			})
			return
		}
		c.JSON(http.StatusOK, Resp{
			Code: 1,
			Msg: "服务端异常，请稍后再试",
		})
		return
	}

	// 登录登录成功，生成token返回给用户
	token,err := GenToken(u.Uid, u.Name)
	if err != nil{
		// 生成token失败
		c.JSON(http.StatusOK, Resp{Code: 1, Msg: "服务端异常，请稍后再试"})
		return
	}

	// 3，返回响应
	c.JSON(http.StatusOK, Resp{
		Code: 0,
		Msg: "success",
		Data: token,
	})

}

// regHandler 注册用户的函数
func regHandler(c *gin.Context){
	var param AuthParam
	// 获取参数，参数解析，根据结构体的tag，来进行校验，绑定到param结构体中
	if err := c.ShouldBind(&param);err != nil{
		c.JSON(http.StatusOK, Resp{
			Code: 1,
			Msg: "参数错误",
		})
		return
	}

	// 校验成功，拿着参数注册用户，数据库中创建一条记录
	var user Account
	// 用name字段因为，建Account表的时候，定义的name是唯一约束，解析前端的参数在param中
	// 所以传param.Name来验证用户是否已经创建过，查出来的数据赋值给 user结构体变量
	err := db.Where("name = ? ", param.Name).First(&user).Error;
		// 错误有2中可能
		// 用户名存在
		if err == nil{
			c.JSON(http.StatusOK, Resp{
				Code: 1,
				Msg: "用户名已存在",	//可以返回参数错误，因为这样容易被暴力破解用户，最好返回其他数据
			})
			return
		}

		// 也不是 查询name 没有注册过导致数据库查询应该是要报 gorm.ErrRecordNotFound 的错
		// 因为取反了，那就是服务端异常
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, Resp{
				Code: 1,
				Msg: "服务端异常, 请稍后再试",	//可以返回参数错误，因为这样容易被暴力破解用户，最好返回其他数据
			})
			return
		}

	// 走到这，代表没有查到name 可以注册用户，也就是报错是 gorm.ErrRecordNotFound
	err = db.Create(&Account{
		Uid: time.Now().Unix(),	// 现在用时间戳生成唯一id ,todo 后面用雪花算法实现唯一的id，
		Name: param.Name,
		Password: md5secret(param.Password),
	}).Error

	if err != nil{
		c.JSON(http.StatusOK, Resp{
			Code: 1,
			Msg: "服务端异常, 请稍后再试",	//可以返回参数错误，因为这样容易被暴力破解用户，最好返回其他数据
		})
		return
	}
	// 没报错，注册成功，让用户登录一遍后再生成token返回
	c.JSON(http.StatusOK, Resp{
		Code: 0,
		Msg: "注册成功",	//可以返回参数错误，因为这样容易被暴力破解用户，最好返回其他数据
	})
	return



}

// md5 加密密码
func md5secret(pwd string) string{
	h := md5.New()
	h.Write([]byte(pwd))
	// 再通过 加盐的 随机字符串进行加密
	return hex.EncodeToString(h.Sum([]byte(MySecret)))
}



