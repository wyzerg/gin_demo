package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin" // gin框架

	"gorm.io/driver/mysql" // gorm 需要导入
	"gorm.io/gorm"         // gorm 需要导入
)

// 结构体 对应数据库的一张表
type Todo struct {
	// gorm.Model
	// 吧 gorm.Model 里面的字段拿出来放在这，然后注释掉 gorm.Model
	// 这样就能给这些字段设置tag: json:"-"，保证返回页面的时候，不返回这些字段
	gorm.Model	// 改成gorm.DeletedAt不然报错，并且添加tag：json:"-" 代表不返回这个字段给前端
	Title  string `form:"title" json:"title"`           // 待办事项名称
	Status bool   `json:"status"`                       //  待办事项 是否完成的状态

	// index 添加索引，关联账户表的 Uid ，不用数据库外键，方便分库分表，只存数据	,因为绝大多数都根据uid来增删改查的，可以增加索引
	Uid    int64  `gorm:"uid;not null;default:0;index"` // 根据这一列能知道是谁的待办事项
}

// Account 用户表
type Account struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Uid int64 `gorm:"uid,unique"` // 设置Uid字段 unique 唯一标识，通过用户名查表

	Name     string `gorm:"name,unique"` // 用户名
	Password string `gorm:"password"`

	NickName string `gorm:"nick_name"` // 昵称随便改
	Status   *bool  `gorm:"status"`
}

var db *gorm.DB // 全局的db对象

func initDB() (err error) {
	dsn := "root:123123@tcp(127.0.0.1:3306)/gogogo?charset=utf8mb4&parseTime=True"
	// 初始化 全局的db对象，所以不用:=，直接用=号获取全局的db，
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return err
}

func main() {
	// 连接数据库
	if err := initDB(); err != nil {
		fmt.Println("initDB connect mysql err:", err)
		panic(err)
	}

	// 使用 Todo结构体(传指针)，来自动创建表
	db.AutoMigrate(&Todo{})
	// 创建用户表
	db.AutoMigrate(&Account{})

	r := gin.Default()
	// 加载前端静态文件 和 static 静态文件返回，并增加页面请求的路由
	r.LoadHTMLFiles("./index.html")
	r.Static("static", "./static")

	// 注册
	r.POST("/register",regHandler)
	// 登录
	r.POST("/login",loginHandler)


	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	// 注册路由，curd
	// 添加待办事项的路由组 g
	g := r.Group("/api/v1", authMiddleware)	// 给路由组添加jwt权限认证中间件
	{
		g.POST("/todo", createTodoHandler)
		g.PUT("/todo", updateTodoHandler)
		g.GET("/todo", getTodoHandler)
		// delete 方式，url是参数在url里面  http://127.0.0.1:8888/api/v1/todo/1，参数赋值给id
		g.DELETE("/todo/:id", deleteTodoHandler)
	}

	fmt.Println("http://127.0.0.1:8888/")
	// 启动http server
	r.Run(":8888")
}



