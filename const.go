package main

const(
	// token中获取uid存到 全局ctx(c)中，包括handler方法从全局ctx获取token接续出来的uid的变量名保持统一
	CtxUidKey = "uid"
	CtxNameKey = "name"
)

