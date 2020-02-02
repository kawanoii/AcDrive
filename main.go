package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/kataras/iris"
	"github.com/kataras/iris/websocket"
)

var status Status

func init() {
	status.Downs = make(map[string]DownStatus)
	status.Ups = make(map[string]UpStatus)
	status = rStatus()
	ck, err := rCookie()
	if err != nil {
		status.AcAuth = 0
	} else {
		_, err := getUpToken(ck)
		if err != nil {
			status.AcAuth = -1
		} else {
			status.AcAuth = 1
		}
	}
}

func main() {
	var ck Cookies
	if status.AcAuth == 1 {
		ck, _ = rCookie()
	}
	wsc := make(chan *websocket.Conn)
	downch := make(chan DownStatus)
	upch := make(chan UpStatus)

	app := iris.New()

	ws := websocket.New(websocket.DefaultGorillaUpgrader, websocket.Events{
		websocket.OnNativeMessage: func(nsConn *websocket.NSConn, msg websocket.Message) error {
			log.Printf("Server got: %s from [%s]", msg.Body, nsConn.Conn.ID())

			nsConn.Conn.Server().Broadcast(nsConn, msg)
			return nil
		},
	})

	go func() {
		for c := range wsc {
			msg, _ := json.Marshal(status)
			c.Socket().WriteText(msg, 0)
			go func(c *websocket.Conn) {
				for {
					select {
					case downs := <-downch:
						status.Downs[downs.Ncd] = downs
					case ups := <-upch:
						status.Ups[ups.Filename] = ups
					}
					msg, _ := json.Marshal(status)
					c.Socket().WriteText(msg, 0)
				}
			}(c)
		}
	}()

	ws.OnConnect = func(c *websocket.Conn) error {
		log.Printf("[%s] Connected to server!", c.ID())
		wsc <- c
		return nil
	}

	ws.OnDisconnect = func(c *websocket.Conn) {
		wStatus(status)
		log.Printf("[%s] Disconnected from server", c.ID())
	}

	app.Favicon("./static/favicon.ico")
	app.Get("/status", websocket.Handler(ws))
	app.Get("/", func(ctx iris.Context) {
		ctx.ServeFile("./web/index.html", false)
	})

	// app.Get("/status", func(ctx iris.Context) {
	// 	//把结构体类型  转成json
	// 	ctx.JSON(status)
	// })

	app.Post("/login", func(ctx iris.Context) {
		var user AcUser
		ctx.ReadForm(&user)

		resp := make(map[string]string)
		ck, err := login(user.Username, user.Password)
		if err != nil {
			status.AcAuth = 0
			resp["code"] = "-1"
			resp["error"] = err.Error()
		} else {
			status.AcAuth = 1
			resp["code"] = "0"
			wCookie(ck)
		}
		ctx.JSON(resp)
	})

	app.Post("/upload", func(ctx iris.Context) {
		var upreq UpReq
		ctx.ReadForm(&upreq)

		var upstatus UpStatus
		if _, ok := status.Ups[upreq.FileName]; ok {
			upstatus = status.Ups[upreq.FileName]
		} else {
			status.Ups[upreq.FileName] = upstatus
		}

		go upload(upreq.FileName, upreq.BlockSize, upreq.Thread, ck, upstatus, upch)
	})

	app.Post("/download", func(ctx iris.Context) {
		var downreq DownReq
		ctx.ReadForm(&downreq)
		if len(downreq.Ncd) < 11 {
			ctx.JSON(map[string]string{"code": "-1", "msg": "链接有误！"})
			return
		}
		if downreq.Thread < 1 {
			downreq.Thread = 4
		}
		var downstatus DownStatus
		if _, ok := status.Downs[downreq.Ncd]; ok {
			downstatus = status.Downs[downreq.Ncd]
		} else {
			downstatus.Ncd = downreq.Ncd
			status.Downs[downreq.Ncd] = downstatus

		}

		go download(downreq.Ncd, downreq.Thread, downstatus, downch)
		ctx.JSON(map[string]string{"code": "0", "msg": "已添加入下载队列！"})

	})

	app.Get("/info/{ncd:string}", func(ctx iris.Context) {
		ncd := ctx.Params().Get("ncd")
		info, _ := infoMeta(ncd)
		ctx.JSON(info)
	})

	app.Get("/exit", func(ctx iris.Context) {
		for key, downs := range status.Downs {
			downs.Code = 3
			status.Downs[key] = downs

		}
		for key, ups := range status.Ups {
			ups.Code = 3
			status.Ups[key] = ups
		}
		wStatus(status)
		os.Exit(0)
	})

	app.Run(iris.Addr(":8080"), iris.WithCharset("UTF-8"))

}
