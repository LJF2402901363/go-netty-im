/*
	@author: 24029

@since: 2023/7/17 23:38:40
@desc:
*/
package main

import (
	"fmt"
	"go-netty-im/common"
	"net/http"

	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty-transport/websocket"
	"github.com/go-netty/go-netty/codec/format"
	"github.com/go-netty/go-netty/codec/frame"
)

var ManagerInst = common.NewManager()

func main() {

	// index page.
	websocket.DefaultOptions.ServeMux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write(indexHtml)
		if err != nil {
			return
		}
	})

	// child pipeline initializer.
	setupCodec := func(channel netty.Channel) {
		channel.Pipeline().
			// read websocket message
			AddLast(frame.PacketCodec(128)).
			// decode bytes to map[string]interface{}
			AddLast(format.JSONCodec(true, false)).
			// session recorder.
			AddLast(ManagerInst).
			// chat handler.
			AddLast(chatHandler{})
	}

	// setup bootstrap & startup server.
	err := netty.NewBootstrap(netty.WithChildInitializer(setupCodec), netty.WithTransport(websocket.New())).
		Listen("0.0.0.0:8080/chat").Sync()
	if err != nil {
		return
	}
}

type chatHandler struct{}

func (chatHandler) HandleActive(ctx netty.ActiveContext) {
	/*if wsTransport, ok := ctx.Channel().Transport().(interface{ HttpRequest() *http.Request }); ok {
		handshakeReq := wsTransport.HttpRequest()
		fmt.Println("websocket header: ", handshakeReq.Header)
	}*/
	fmt.Printf("child connection from: %s\n", ctx.Channel().RemoteAddr())
	ctx.HandleActive()
}

func (chatHandler) HandleRead(ctx netty.InboundContext, message netty.Message) {

	fmt.Printf("received child message from: %s, %v\n", ctx.Channel().RemoteAddr(), message)

	if cmd, ok := message.(map[string]interface{}); ok {
		cmd["id"] = ctx.Channel().ID()
	}

	ManagerInst.Broadcast(message)
}

func (chatHandler) HandleInactive(ctx netty.InactiveContext, ex netty.Exception) {
	fmt.Printf("child connection closed: %s %s\n", ctx.Channel().RemoteAddr(), ex.Error())
	ctx.HandleInactive(ex)
}
