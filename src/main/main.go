package main

import (
	"fmt"
	"container/list"
)

var iotlist list.List		//物联网客户单列表

func delClient(client *ClientInfo){
	del: for e := iotlist.Front(); e != nil; e = e.Next() {
		if(e.Value == client){
			iotlist.Remove(e)
			break del
		}
	}
}

func main() {
	fmt.Println("Start")
	ch := make(chan IotEvent, 10)
	go Server(ClientType_IOT, ":7890", ch)
	go Server(ClientType_CTL, "7891", ch)
	for x := range ch{
		//处理各种收到的事件
		fmt.Println(x)
		evtType := x.EventType
		client := x.Client
		if(evtType == NewClient){
			if(client.ClientType == ClientType_IOT){
				iotlist.PushBack(client)
			}
		}else if(evtType == DisConn){
			//客户端断开
			if(client.ClientType == ClientType_IOT){
				delClient(client)
			}
		}else if(evtType == ReceiveData){
			//收到数据
			processRecvData(client, x.Data)
		}else{
			//不支持的事件
			fmt.Println("收到不支持的数据类型：", evtType)
		}
	}
}

func disConn(client *ClientInfo){
	fmt.Println("disconn")
	protoBuffer = protoBuffer[0:0]//清空协议分析缓冲区
}

var protoBuffer = make([]byte, 0, 1024)

func processRecvData(client *ClientInfo, data []byte){
	fmt.Println("收到数据长度：", len(data))
	protoBuffer = append(protoBuffer, data...)
	if(len(protoBuffer) < 3){
		//数据不够
		return
	}	
	if(protoBuffer[0] != 0x36 || protoBuffer[1] != 0x50){
		//header fail, shoud disconn
		disConn(client)
		return
	}
	
	datalen := int(protoBuffer[2])
	if((datalen + 3) > len(protoBuffer)){
		//数据不够
		return
	}
	//数据足够一帧消息的
	msg := protoBuffer[3:datalen + 3]
	processMsg(client, msg)
	remain := protoBuffer[datalen + 3:]
	copy(remain, protoBuffer)
}

func sendBackHeartBeat(client *Client){
	
}

func processMsg(client *ClientInfo, msg []byte){
	type := int(byte[0])
	if(type == 0){
		//心跳
		sendBackHeartBeat(client)
	}else if(type == 1){
		
	}
}
