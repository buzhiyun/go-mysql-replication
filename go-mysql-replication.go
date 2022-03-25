package main

import (
	"flag"
	"fmt"
	"github.com/buzhiyun/go-mysql-replication/config"
	"github.com/buzhiyun/go-mysql-replication/controller"
	"github.com/buzhiyun/go-mysql-replication/prometheus"
	"github.com/buzhiyun/go-mysql-replication/service"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var (
	helpFlag bool
	cfgPath  string
	//stockFlag    bool
	positionFlag bool
	//statusFlag   bool
)

func newApp() (app *iris.Application) {
	app = iris.New()

	// go get -u github.com/go-bindata/go-bindata/...
	// 静态文件直接打包到程序里  先执行 go-bindata -fs -nomemcopy -prefix "web/dist" ./web/dist/...
	// https://docs.iris-go.com/iris/file-server/http2push-embedded-compression
	// irirs v12.2.0-alpha 的方法
	//var opts = iris.DirOptions{
	//	IndexName: "index.html",
	//	PushTargetsRegexp: map[string]*regexp.Regexp{
	//		"/": iris.MatchCommonAssets,
	//	},
	//	ShowList: true,
	//	Cache: iris.DirCacheOptions{
	//		Enable:         true,
	//		CompressIgnore: iris.MatchImagesAssets,
	//		Encodings:      []string{"gzip", "deflate", "br", "snappy"},
	//		// Compress files equal or larger than 50 bytes.
	//		CompressMinSize: 50,
	//		Verbose:         1,
	//	},
	//}
	//app.HandleDir("/", AssetFile(), opts)

	// https://github.com/kataras/iris/blob/v12.1.8/_examples/file-server/embedding-files-into-app/main.go
	// go-bindata ./assets/...
	app.HandleDir("/", "./assets", iris.DirOptions{
		Asset:      Asset,
		AssetInfo:  AssetInfo,
		AssetNames: AssetNames,
	})

	//测试的时候允许跨域
	Cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // 这里写允许的服务器地址，* 号标识任意
		AllowCredentials: true,
	})
	api := app.Party("/api", Cors).AllowMethods(iris.MethodOptions)
	//api := app.Party("/api").AllowMethods(iris.MethodOptions)

	api.PartyFunc("/canal", func(server iris.Party) {
		server.Get("/state", controller.GetCanalState)
		server.Post("/start", controller.StartCanal)
		server.Post("/stop", controller.StopCanal)
		server.Get("/info", controller.GetCanalInfo)
	})
	api.PartyFunc("/transfer", func(server iris.Party) {
		server.Get("/endpoint", controller.GetEndpointType)
		server.Get("/endpoint/state", controller.GetEndpointState)
		server.Get("/state", controller.GetTransferState)
		server.Post("/start", controller.StartTransferManual)
	})
	api.PartyFunc("/report", func(server iris.Party) {
		server.Get("/table", controller.GetTableReport)
		server.Get("/schema", controller.GetSchemaReport)
	})
	app.Get("/metric", iris.FromStd(promhttp.Handler()))

	return
}

func main() {

	flag.BoolVar(&helpFlag, "help", false, "this help")
	flag.StringVar(&cfgPath, "c", "config.yml", "config file")
	//flag.BoolVar(&stockFlag, "stock", false, "stock data import")
	flag.BoolVar(&positionFlag, "position", false, "set dump position")
	//flag.BoolVar(&statusFlag, "status", false, "display application status")
	flag.Usage = usage

	flag.Parse()

	// 查看help
	if helpFlag {
		flag.Usage()
		return
	}

	golog.SetTimeFormat("2006-01-02 15:04:05")
	//golog.SetLevel("DEBUG")

	// 初始化global
	err := config.Initialize(cfgPath)
	if err != nil {
		golog.Error(err)
		return
	}

	service.Start()

	if err := prometheus.Initialize(); err != nil {
		golog.Error(err)
		return
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Kill, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// 打开web
	if config.GlobalConfig.EnableWebAdmin {
		app := newApp()
		app.Run(iris.Addr("0.0.0.0:" + strconv.Itoa(config.GlobalConfig.WebAdminPort)))
	}

	sin := <-s

	golog.Infof("application stoped，signal: %s \n", sin.String())

	service.Close()
}

func usage() {
	fmt.Fprintf(os.Stdout, `version: 1.0.0
Usage: go-mysql-replication [-c configfile]

Options:
`)
	flag.PrintDefaults()
}
