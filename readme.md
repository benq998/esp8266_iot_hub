### 协议定义
* 两个字节magic head  
`0x36 0x50`
* 一个直接数据长度
* 一个直接消息类型  
	```
	0-心跳报文，一般有iot端发起，hub端回应即可
	1-hub向iot转发控制消息
	2-iot向hub发上行消息
	10-ctl向hub请求iot端列表
	11-ctl想hub发控制消息
	```
* 数据
