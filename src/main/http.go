package main

import (
	"fmt"
	http "net/http"
)

var keyMap = map[string][]byte{
	"sr_on": {0xFD,0x01,0x01,0xAE,0xBF,0xF4,0x56,0xDF},//sr标识small room
	"sr_off":{0xFD,0x01,0x01,0xAE,0xBF,0xF2,0x56,0xDF},
}

func startHttp(){
	http.HandleFunc("/switch", IndexHandler)
    http.ListenAndServe(":7892", nil)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	which := r.FormValue("which")
	fmt.Println("which:",which)
	var val []byte
	loop: for k,v := range keyMap {
		if(k == which){
			val = v
			break loop
		}
	}
	if(val != nil){
		exec(val)
		fmt.Fprintln(w, "ok")
	}else{
	    fmt.Fprintln(w, "err")
	}
}

func exec(data []byte){
	fmt.Println("sendCtlData:", data)
	sendCtlData(data)
}
