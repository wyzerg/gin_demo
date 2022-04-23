package main


// 小清单的增删改查

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// createTodoHandler 创建
func createTodoHandler(c *gin.Context) {
	// 3个步骤

	// 1，获取参数 取title字段，比如前端传入json数据 {"title":"计划1"}
	var todo Todo // 尝试将数据解析到 todo中去，需要将结构体设置json的tag
	if err := c.ShouldBind(&todo); err != nil {
		fmt.Println("createTodoHandler 获取参数错误：", err)
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "无效的参数", // 正常情况，不能直接返回错误给前端
		})
		return // 错误就不往后走
	}

	// c 是贯穿gin框架所有函数，通过c获取当前用户的Uid，中间件中通过c 吧uid传到 ctx中了，
	v,_ := c.Get(CtxUidKey)
	// 拿到数据断言是否是int64类型
	uid := v.(int64)
	if uid <= 0{
		// 异常
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "登录异常", // 正常情况，不能直接返回错误给前端
		})
		return
	}

	todo.Uid = uid
	// 2，处理业务逻辑，新增一条数据
	if err := db.Create(&todo).Error; err != nil {
		fmt.Println("db.Create err：", err)
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "服务端异常", // 正常情况，不能直接返回错误给前端
		})
		return // 错误就不往后走
	}

	// 3，返回响应
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		//"data":todo,	// 也可以返回数据给前端
	})
}

// updateTodoHandler 修改
func updateTodoHandler(c *gin.Context) {
	// 1，获取参数
	var todo Todo
	if err := c.ShouldBind(&todo); err != nil {
		fmt.Println("updateTodoHandler err：", err)
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "无效的参数",
		})
		return
	}

	// 2，执行业务逻辑，更新数据

	// 2.1 根据从c中获取uid，中间件传的uid，来获取对应uid下的待办事项
	v,_ := c.Get(CtxUidKey)
	// 接口类型要断言
	uid := v.(int64)
	if uid <= 0{
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "登录异常，请重新登录", // 正常情况，不能直接返回错误给前端
		})
		return
	}
	//// 先将前端传来赋值给 todo对象的id属性， 现在数据库中查询一遍，然后吧数据保存到obj中
	// 这里注释的原因是：下面 吧 uid 和 todo.ID 作为联合条件查询，所以这里不需要判断了
	//if todo.Uid != uid{
	//	// 修改的数据是不是自己的数据
	//	c.JSON(200, gin.H{
	//		"code": 1,
	//		"msg":  "登录异常，请重新登录", // 正常情况，不能直接返回错误给前端
	//	})
	//	return
	//}


	// 2.2 先检查数据是否存在(根据主键检索)，因为在执行 c.ShouldBind之后，自动吧前端的数据填充到结构体中
	// 参考：https://gorm.io/zh_CN/docs/query.html#%E7%94%A8%E4%B8%BB%E9%94%AE%E6%A3%80%E7%B4%A2
	// 这里吧 uid 和 todo.ID 作为联合条件查询
	var obj Todo
	if err := db.Where("id = ? and uid = ?",todo.ID, uid).First(&obj).Error; err != nil {
		// 返回2种错误的第1种： 没有这条记录的错误，通过 errors.Is 断言递归查找错误类型是否是 gorm.ErrRecordNotFound
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(200, gin.H{
				"code": 1,
				"msg":  "无效的参数",
			})
			return
		}
		// 返回2种错误的第2种： 其他错误
		fmt.Println("updateTodoHandler db.Save err：", err)
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "服务端异常，请稍后再试", // 正常情况，不能直接返回错误给前端
		})
		return
	}


	// 验证过有数据后走到这，更新指定的字段"status",.Debug()显示具体执行的sql语句, 例如前端传 {"id":2,"status": true}
	if err := db.Debug().Model(&todo).Update("status", todo.Status).Error; err != nil {
		//if err := db.Save(&todo);err != nil{
		// db.Save更新所有字段，因为gin框架创建的表有很多其他默认带的字段，前端传过来没有gin框架生成的字段和字段的值
		// 这样会因为gin框架默认字段没有值而报错，所以选择更新指定字段，用 db.Model(&todo).Update("status", todo.Status)
		fmt.Println("updateTodoHandler db.Save err：", err)
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "服务端异常，请稍后再试", // 正常情况，不能直接返回错误给前端
		})
		return
	}

	// 3，返回响应
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
	})
}

// getTodoHandler 查询所有待办事项
func getTodoHandler(c *gin.Context) {
	// 1，获取请求参数，由于返回所有数据，就不获取参数
	// 2，执行业务逻辑
	// 根据请求的uid获取全部对象 db.find()，参考 https://gorm.io/zh_CN/docs/query.html#%E6%A3%80%E7%B4%A2%E5%85%A8%E9%83%A8%E5%AF%B9%E8%B1%A1
	var todos []Todo // todos是Todo类型的切片，如果查询单条数据就是结构体
	// 2.1 根据从c中获取uid，中间件传的uid，来获取对应uid下的待办事项
	v,_ := c.Get(CtxUidKey)
	// 接口类型要断言
	uid := v.(int64)
	if uid <= 0{
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "登录异常，请重新登录", // 正常情况，不能直接返回错误给前端
		})
		return
	}

	// 根据token解析出来的uid查询
	if err := db.Where("uid = ?",uid).Debug().Find(&todos).Error; err != nil {
		fmt.Println("getTodoHandler 查询失败:", err)
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "服务端异常，请稍后再试", // 正常情况，不能直接返回错误给前端
		})
		return
	}

	// 3，返回响应
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		"data": todos, // 返回的大量字段，有不想返回给前端的，需要添加tag，修改结构体字段后面 json:"-"
	})
}

// deleteTodoHandler 删除
func deleteTodoHandler(c *gin.Context) {
	// 获取参数
	// 前端执行delete方式，url是参数在url里面  http://127.0.0.1:8888/api/v1/todo/1
	idStr := c.Param("id") // 从 路由的 /todo/:id 获取id
	// 防止sql注入，需要将字符串转换int，看是否出错
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("deleteTodoHandler err:", err)
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "无效的参数",
		})
	}

	// 业务逻辑
	// 2.1 根据从c中获取uid，中间件传的uid，来获取对应uid下的待办事项
	v,_ := c.Get(CtxUidKey)
	// 接口类型要断言
	uid := v.(int64)
	if uid <= 0{
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "登录异常，请重新登录", // 正常情况，不能直接返回错误给前端
		})
		return
	}

	// 2.2先根据这个id查一下有没有这条记录
	// 这里改为联合条件查询
	var obj Todo
	if err := db.Debug().Where("id = ? and uid = ?", id, uid).First(&obj).Error; err != nil {
		// 返回2种错误的第1种： 没有这条记录的错误，通过 errors.Is 断言递归查找错误类型是否是 gorm.ErrRecordNotFound
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(200, gin.H{
				"code": 1,
				"msg":  "无效的参数",
			})
			return
		}
		// 返回2种错误的第2种： 其他错误
		fmt.Println("deleteTodoHandler db.First err：", err)
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "服务端异常，请稍后再试", // 正常情况，不能直接返回错误给前端
		})
		return
	}

	// 根据主键删除数据，参考 https://gorm.io/zh_CN/docs/delete.html#%E6%A0%B9%E6%8D%AE%E4%B8%BB%E9%94%AE%E5%88%A0%E9%99%A4
	// 删除是 软删除，给删除的字段添加标记，代表删除，但是数据还在数据库中，只是返回前端代表没有这个数据了
	if err := db.Delete(&obj, id).Error; err != nil {
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "服务端异常，请稍后再试", // 正常情况，不能直接返回错误给前端
		})
		return
	}

	// 返回响应
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
	})
}

