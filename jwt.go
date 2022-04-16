package main

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// 登录成功之后来生成jwt token返回,生成jwt的相关代码

// 方便测试出效果，设置token的有效期：10秒过期，例子中设置2小时过期
const TokenExpireDuration = time.Second * 10

var MySecret = []byte("夏天夏天悄悄过去") // 加盐的密钥

type MyClaims struct {
	// 定义生成token的字段，这些定义的校验的字段，一定在数据库中创建的时候，保证唯一性 unique
	Uid  int64  `json:"uid"`
	Name string `json:"name"`

	jwt.StandardClaims
}

// GenToken 定义登录成功之后(用户名/密码...)，生成JWT的方法，传进来的 uid 和 name就是 MyClaims 结构体的2个校验字段
func GenToken(uid int64, name string) (string, error) {
	// 创建一个我们自己的声明
	c := MyClaims{
		uid,
		name, // 自定义字段
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(), // 过期时间
			Issuer:    "todo-app",                                 // 标识一下签发人
		},
	}
	// 使用指定的签名方式 jwt.SigningMethodHS256 对 对象 进行签名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 使用指定的secret 加盐的字符串进行签名 并获得 完整的 编码后的 字符串token
	return token.SignedString(MySecret)
}


// ParseToken 用来 每次用户请求后端过来，携带token的时候，对token进行解析
func ParseToken(tokenString string) (*MyClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		// 回调函数，把加盐的字符串通过这个函数返回
		return MySecret, nil
	})
	if err != nil {
		return nil, err
	}
	// 解析出的token，如果是之前声明的token，并且 在有效期内(token.Valid)
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid { // 校验token
		// token校验成功
		return claims, nil
	}
	// token校验失败
	return nil, errors.New("invalid token")
}
