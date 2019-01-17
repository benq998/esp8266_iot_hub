package main

import (
	"fmt"
	"net"
	"time"
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

func (c *ClientInfo) SendData(data []byte){
	n,_ := c.conn.Write(data)
	fmt.Println("写出数据：", data,",长度：",n)
}

func Server(clientType ClientTypeVal,laddr string, ch chan IotEvent){
	fmt.Println("OK")
	ln, err := net.Listen("tcp", laddr)
	if err != nil {
		fmt.Println("服务器启动失败")
		return
	}
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
	fmt.Println("OK")
	client := ClientInfo{conn:conn,ClientType:clientType,ConnTime:time.Now().Unix()}
	evt := IotEvent{EventType:NewClient,Client:&client}
	ch <- evt
	go readSocket(&client, ch)
}

func readSocket(client *ClientInfo, ch chan IotEvent){
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