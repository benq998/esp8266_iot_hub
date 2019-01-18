## 看到的大侠别嘲笑，学习中，很多都是简易的实现。

### 协议定义
* 两个字节magic head  
`0x36 0x50`
* 一个字节数据长度，包括后面的消息类型后其他所有数据，不包括当前字段和前面0x36以及0x50
* 一个字节消息类型  
	```
	0-心跳报文，一般有iot端发起，hub端回应即可
	1-hub向iot转发控制消息
	2-iot向hub发上行消息，目前就是串口回应的数据本身，没有特定格式
	10-ctl向hub请求iot端列表
	11-ctl向hub发控制消息，消息本身，没有特定格式
	100-hub回复ctl iot列表信息
	101-hub回复ctl转发控制数据的结果
	```
* 数据

##### 消息100数据格式
* 一个字节客户端数量，后面是每个客户端的数据连续输出
* 四个字节客户端IP
* 两个字节客户端端口
* 四个字节客户端连接到hub上的时间戳，19700101到现在的秒数

##### 消息101数据格式
* 0-转发成功，1-转发失败

### 端口配置
* 7890-iot 设备连接
* 7891-手机控制端连接