package main

import (
	"fmt"
	"container/list"
	"bytes"
	"time"
)

//物联网客户单列表
var iotlist = list.New()

func delClient(client *ClientInfo){
	del: for e := iotlist.Front(); e != nil; e = e.Next() {
		if(e.Value.(*ClientInfo) == client){
			iotlist.Remove(e)
			break del
		}
	}
}

func breakHeartBeatTimeOutClients(){
	for{
		now := int32(time.Now().Unix())
		for e:= iotlist.Front(); e != nil; e = e.Next() {
			client := e.Value.(*ClientInfo)
			if((now - client.LastHeartBeatTime) > 30){//超过30秒强制断开
				client.forceDisconnect();
				iotlist.Remove(e)
				fmt.Println("强制断开心跳超时的设备链接:",client.GetRemoteAddress())
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func main() {
	fmt.Println("Start")
	ch := make(chan IotEvent, 10)
	go Server(ClientType_IOT, "0.0.0.0:7890", ch)
	go Server(ClientType_CTL, "0.0.0.0:7891", ch)
	go breakHeartBeatTimeOutClients()
	go startHttp()
	for x := range ch{
		//处理各种收到的事件
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
	fmt.Printf("收到数据(%d)：%s\n", len(data), hexToString(data))
	protoBuffer = append(protoBuffer, data...)
//	fmt.Printf("和已有数据连接后(%d)：%s\n", len(protoBuffer), hexToString(protoBuffer))
	loop: for{
		if(len(protoBuffer) < 3){
			//数据不够
			break loop
		}	
		if(protoBuffer[0] != 0x36 || protoBuffer[1] != 0x50){
			//header fail, shoud disconn
			disConn(client)
			break loop
		}
		
		datalen := int(protoBuffer[2])
		if((datalen + 3) > len(protoBuffer)){
			//数据不够
			break loop
		}
		//数据足够一帧消息的
		msg := protoBuffer[3:datalen + 3]
		processMsg(client, msg)
		remain := protoBuffer[datalen + 3:]
		copy(remain, protoBuffer)
		protoBuffer = protoBuffer[:len(remain)]
	}
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
//	fmt.Println("收到心跳：", msg)
	client.LastHeartBeatTime = int32(time.Now().Unix())
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
	outBuf = append(outBuf, client.GetAddressAsBytes()...)
	outBuf = append(outBuf, client.GetConnTimeAsBytes()...)
	return outBuf
}

//转发控制数据,ctlData里只是控制数据，没有消息类型
func forwardCtlMsg(client *ClientInfo, ctlData []byte){
	iotCnt := iotlist.Len()
	forwardRst := []byte{101}
	if(iotCnt > 0){
		sendCtlData(ctlData)
		//转发完成
		fmt.Println("转发控制消息完成")
		forwardRst = append(forwardRst, 0)
	}else{
		//没有客户端，无法转发
		fmt.Println("还没有iot设备，无法转发控制消息")
		forwardRst = append(forwardRst, 1)
	}
	sendProtocol(client, forwardRst)
}

func sendCtlData(data []byte){
	data = append([]byte{1}, data...)
	for iot:=iotlist.Front();iot!=nil;iot = iot.Next() {
		sendProtocol(iot.Value.(*ClientInfo), append([]byte{1}, data...))
	}
}

func processMsg(client *ClientInfo, msg []byte){
	//fmt.Println(hexToString(msg))
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

func hexToString(data []byte)string{
	var buf bytes.Buffer
	fmt.Fprintf(&buf,"HEX:")
	for x:=range data{
		if(x > 0){
			fmt.Fprintf(&buf," %02X", data[x])
		}else{
			fmt.Fprintf(&buf,"%02X", data[x])
		}
	}
	return buf.String()
}