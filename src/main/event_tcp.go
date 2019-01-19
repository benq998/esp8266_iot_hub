package main

import (
	"fmt"
	"net"
	"time"
	"bytes"
	reg "regexp"
	strv "strconv"
	binary "encoding/binary"
)

type EventTypeVal int

const(
	NewClient EventTypeVal = 0
	DisConn EventTypeVal = 1
	ReceiveData EventTypeVal = 2
)

type ClientTypeVal int
const(
	ClientType_IOT ClientTypeVal = 0
	ClientType_CTL ClientTypeVal = 1
)

type IotEvent struct{
	EventType EventTypeVal		//事件类型 ref EventTypeVal
	Client *ClientInfo	//连接对象
	Data []byte			//数据缓冲区
}

type ClientInfo struct{
	conn net.Conn
	ClientType ClientTypeVal		//客户端类型，0-iot客户端，1-控制端
	ConnTime int64
}

//返回4字节ip，2字节端口
func (c *ClientInfo) GetAddressAsBytes()(ipport []byte){
	buf := new(bytes.Buffer)
	addr := c.conn.RemoteAddr()
//	fmt.Println("addr:",addr.String())
	part,_ := reg.Compile("\\.|:")
	ips := part.Split(addr.String(), -1)
//	fmt.Println("ips:",ips)
	for x := range ips{
		if(x <= 3){
			b,_ := strv.Atoi(ips[x])
			fmt.Println(ips[x],"=>",b)
			_ = binary.Write(buf, binary.BigEndian, byte(b))
		}else{
			s,_ := strv.Atoi(ips[x])
			fmt.Println(ips[x],"=>",s)
			_ = binary.Write(buf, binary.BigEndian, int16(s))
		}
	}
	return buf.Bytes()
}

//返回4字节的秒数
func (c *ClientInfo) GetConnTimeAsBytes()[]byte{
	buf := new(bytes.Buffer)
//	fmt.Println("c.ConnTime:", c.ConnTime)
	err := binary.Write(buf, binary.BigEndian, int32(c.ConnTime))
	if(err == nil){
		return buf.Bytes()
	}else{
		return []byte{0,0,0,0}
	}
}

func (c *ClientInfo) SendData(data []byte){
	n,_ := c.conn.Write(data)
	fmt.Println("写出数据：", data,",长度：",n)
}

func Server(clientType ClientTypeVal,laddr string, ch chan IotEvent){
	fmt.Println("start server addr:", laddr)
	ln, err := net.Listen("tcp4", laddr)
	if err != nil {
		fmt.Println("服务器启动失败")
		return
	}
	fmt.Println("监听成功，等待客户端连接")
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			continue
		}
		go handleConnection(clientType,conn, ch)
	}
}

func handleConnection(clientType ClientTypeVal, conn net.Conn, ch chan IotEvent){
	fmt.Println("new conn:", conn.RemoteAddr().String())
	client := ClientInfo{conn:conn,ClientType:clientType,ConnTime:time.Now().Unix()}
	evt := IotEvent{EventType:NewClient,Client:&client}
	ch <- evt
	go readSocket(&client, ch)
}

func readSocket(client *ClientInfo, ch chan IotEvent){
	fmt.Println("readSocket:", client.conn.RemoteAddr().String())
	buf := make([]byte, 1024)
	loop: for{
		n, err := client.conn.Read(buf)
		if(err != nil){
			break loop
		}
		data := buf[:n]
		ch <- IotEvent{EventType:ReceiveData,Client:client, Data:data}
	}
	ch <- IotEvent{EventType:DisConn,Client:client}
}