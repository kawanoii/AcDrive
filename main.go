package main

import "github.com/kataras/iris"

var status Status

func init() {
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
	app := iris.New()
	app.RegisterView(iris.HTML("./static", ".html").Reload(true))
	app.Favicon("./static/favicon.ico")

	app.Get("/", func(ctx iris.Context) {
		ctx.View("index.html")
	})

	app.Get("/status", func(ctx iris.Context) {
		//把结构体类型  转成json
		ctx.JSON(status)
	})

	app.Post("/login", func(ctx iris.Context) {
		var user AcUser
		ctx.ReadJSON(&user)
		ck, err := login(user.Username, user.Password)
		if err != nil {
			//返回错误
			return
		}
		wCookie(ck)
	})

	app.Post("/upload", func(ctx iris.Context) {
		var upreq UpReq
		ctx.ReadJSON(&upreq)

		var upstatus UpStatus
		status.Ups = append(status.Ups, upstatus)
		// 请求参数格式化  请求参数是json类型转化成 UpReq 类型
		// 比如 post 参数 {filepath:'xxxx'} 转成 UpReq 类型
		//把 json 类型请求参数 转成结构体
		go upload(upreq.FilePath, upreq.BlockSize, upreq.Thread, ck, upstatus)
	})

	app.Post("/download", func(ctx iris.Context) {
		var downreq DownReq
		ctx.ReadJSON(&downreq)

		var downstatus DownStatus
		status.Downs = append(status.Downs, downstatus)
		go download(downreq.Ncd, downreq.Thread, downstatus)
	})

	app.Get("/info", func(ctx iris.Context) {
		// ncd := ctx.Params().Get("ncd")
		ncd := "nCoVDrive://aHR0cHM6Ly9pbWdzLmFpeGlmYW4uY29tL21ldGFfZTY5MmJkODRkOTgxMTUyNGQ2OTZhZjVhMWJjODlhZDdjZTQ3OWQwMw=="
		info, err := infoMeta(ncd)
		if err != nil {
			//返回错误

		}
		ctx.JSON(info)
	})

	app.Get("/profile/{username:string}", profileByUsername)
	app.Run(iris.Addr(":8080"), iris.WithCharset("UTF-8"))

}
func profileByUsername(ctx iris.Context) {
	//获取路由参数
	username := ctx.Params().Get("username")
	//向数据模板传值 当然也可以绑定其他值
	ctx.ViewData("Username", username)
	//渲染模板 ./web/views/profile.html

	//把获得的动态数据username 绑定在 ./web/views/profile.html 模板 语法{{}} {{ .Username }}

	ctx.View("profile.html")
}
