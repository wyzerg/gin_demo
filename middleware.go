package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// 中间件


// ctxKey 防止保存c上下文的属性和名称冲突，创造一个自定义类型
// 但是 不涉及夸包调用，不会吧gin框架的ctx传给其他包，所以可以这样写
// 在 Context.Context 中才会用到
//type ctxKey string
//var (
//	KeyUid ctxKey = "uid"	// 根据自定义类型声明一个变量
//	KeyName ctxKey = "name"
//)


// authMiddleware 从请求头中获取 token，完成校验
func authMiddleware(c *gin.Context){
	// 1，从请求 头中获取token
	// token 一般放在请求头(jwt推荐)，或者请求体，或者url参数中

	// c.Request 是原始的请求request,从头部的 Authorization 字段中获取token, Authorization 字段是前后端商量好自定义的
	authHeader := c.Request.Header.Get("Authorization")
	//authHeader := c.Request.Header.Get("token")		// 也可以自定义前端的key是什么
	//fmt.Println(authHeader)


	if authHeader == ""{
		c.JSON(http.StatusOK, Resp{
			Code: 1,
			Msg: "请求头 Bearer Token为空",
		})
		c.Abort()	// 终止函数，不跳转到下面的函数了，直接返回
		return
	}

	// 2，解析token
	parts := strings.SplitN(authHeader, " ", 2)	// 按照空格切割，分成2段
	// 拿到token，如果token获取到的是 Bearer tokenxxx 需要切割
	// 如果直接是token，就不用切割
	if !(len(parts) == 2 && parts[0] == "Bearer"){
		c.JSON(http.StatusOK, Resp{
			Code: 1,
			Msg: "请求头Bearer auth格式错误",
		})
		c.Abort()	// 终止函数，不跳转到下面的函数了，直接返回
		return
	}

	// 3，校验token
	// 走到这，拿到了正确的token 在切割的索引1的切片中
	mc,err := ParseToken(parts[1])
	if err != nil{
		c.JSON(http.StatusOK, Resp{
			Code: 1,
			Msg: "无效的Token",
		})
		c.Abort()	// 终止函数，不跳转到下面的函数了，直接返回
		return
	}
	// TODO  还可以添加，解析token成功后，从redis 根据userid查， 存redis的步骤在auth.go中，用户登录成功后生成token之后，就存redis


	// 将当前请求的 Name和 Uid 信息保存到请求的上下文 c 上
	// 怕和原始的类型冲突，可以定义自定义的类型

	// c.Set("name", mc.Name)  如果这样写，项目其他人有可能也会填共同样的字段保存，就会互相覆盖，不想互相覆盖，就设置自定义的类型，存储进去
	// 但是 不涉及夸包调用，不会吧gin框架的ctx传给其他包，所以可以这样写， 在 Context.Context 中才会用到
	c.Set(CtxNameKey, mc.Name)
	c.Set(CtxUidKey, mc.Uid)

	c.Next()	// 最后一步 ，可以写 next，也可以不写next，都会跳转到下一个函数
}

