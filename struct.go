package main

// Rest 约定好后端api，统一响应的结构体
type Resp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`	// omitempty代表没有data的话，就省略掉
}
