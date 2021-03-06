package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
	"io"
	//"os"
	//"io/ioutil"
	//"io/ioutil"
	"bufio"
	"net/http"
	"os/exec"
	"strings"
)

// Operations about Users
type TermController struct {
	beego.Controller
}

var containerid = "null"

var wsmap_term = make(map[string]*websocket.Conn)

// @Title testterm
// @Description : start the websocket connection
// @Param	body		body 	models.User	true		"body for user content"
// @Success 200 {int} models.User.Id
// @Failure 403 body is empty
// @router / [get]
func (o *TermController) Get() {
	endpoint := o.Ctx.Request.RemoteAddr
	url := strings.Split(endpoint, ":")[0]
	fmt.Println(url)
	ws, err := websocket.Upgrade(o.Ctx.ResponseWriter, o.Ctx.Request, nil, 1024, 1024)
	wsmap_term[url] = ws
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(o.Ctx.ResponseWriter, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		beego.Error("Cannot setup WebSocket connection:", err)
		return
	}
	o.Ctx.WriteString("connection ok")

	//start the pty
	// ubuntu:latest
	c := exec.Command("docker", "run", "-it", "ubuntu:latest", "/bin/bash")
	//c := exec.Command("/bin/bash")
	f, err := pty.Start(c)
	if err != nil {
		panic(err)
	}
	//pipeReader, pipeWriter := io.Pipe()
	wsm := wsmap_term[url]
	go func() {

		for {
			_, p, err := wsm.ReadMessage()
			if err != nil {
				panic(err)
			}

			fmt.Println(string(p))
			p = append(p, 10)

			io.Copy(f, strings.NewReader(string(p)))

		}
	}()
	getid := true
	var id = "null"
	//buffer := make([]byte, 1000)
	go func() {

		for {

			r := bufio.NewReader(f)
			line, _, err := r.ReadLine()
			if err != nil {
				break
			}
			if getid {
				str1 := strings.Split(string(line), "@")[1]
				str2 := strings.Split(str1, ":")[0]
				containerid = str2
				getid = false
			}

			fmt.Println("id:", id)
			line = append(line, 10)
			fmt.Println("start", string(line), "end")
			for {
				err = wsm.WriteMessage(websocket.TextMessage, []byte(line))
				if err == nil {
					break
				}

			}

		}
	}()

}
