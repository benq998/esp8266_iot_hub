package main

import (
	"fmt"
	"container/list"
)

//物联网客户单列表
var iotlist = list.New()

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
	go Server(ClientType_CTL, ":7891", ch)
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

func sendProtocol(client *ClientInfo, msg []byte){
	msgLen := len(msg)
	outMsg := make([]byte, msgLen + 3)
	outMsg[0] = 0x36
	outMsg[1] = 0x50
	outMsg[2] = byte(msgLen)
	copy(outMsg[3:], msg)
	client.SendData(outMsg)
}

//回复iot心跳
func sendBackHeartBeat(client *ClientInfo,msg []byte){
	fmt.Println("收到心跳：", msg)
	sendProtocol(client, msg)
}

//下发iot列表信息
func sendIotList(client *ClientInfo){
	msg := make([]byte,0,256)
	iotCount := byte(iotlist.Len())
	msg = append(msg, 100, iotCount)
	for e := iotlist.Front(); e != nil; e = e.Next(){
		msg = writeIotClientInfo(msg, e.Value.(*ClientInfo))
	}
	sendProtocol(client, msg)
}

func writeIotClientInfo(outBuf []byte, client *ClientInfo)[]byte{
	ipport := client.GetAddressAsBytes()
	outBuf = append(outBuf, ipport...)
	outBuf = append(outBuf, client.GetConnTimeAsBytes()...)
	return outBuf
}

//转发控制数据,ctlData里只是控制数据，没有消息类型
func forwardCtlMsg(client *ClientInfo, ctlData []byte){
	iotCnt := iotlist.Len()
	forwardRst := []byte{101}
	if(iotCnt > 0){
		iot := iotlist.Front().Value.(*ClientInfo)
		sendProtocol(iot, append([]byte{1}, ctlData...))
		//转发完成
		forwardRst = append(forwardRst, 0)
	}else{
		//没有客户端，无法转发
		forwardRst = append(forwardRst, 1)
	}
	sendProtocol(client, forwardRst)
}

func processMsg(client *ClientInfo, msg []byte){
	msgType := int(msg[0])
	if(msgType == 0){
		//心跳
		sendBackHeartBeat(client, msg)
	}else if(msgType == 2){
		//iot to hub
		fmt.Println("iot data:", msg)
	}else if(msgType == 10){
		//ctl 向hub请求iot列表
		sendIotList(client)
	}else if(msgType == 11){
		//ctl 向 hub发控制消息
		forwardCtlMsg(client,msg[1:])
	}else{
		//不支持的消息类型
		fmt.Println("不支持的消息类型：", msgType)
	}
}
