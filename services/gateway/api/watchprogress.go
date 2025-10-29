package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"watch-progress-service/services/gateway/api/internal/config"
	"watch-progress-service/services/gateway/api/internal/handler"
	"watch-progress-service/services/gateway/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

func main() {
	configFile := flag.String("config", "services/gateway/api/etc/watchprogress.yaml", "config file path")
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.Rest, rest.WithFileServer("/api/doc", http.Dir("./services/gateway/api/docs")))
	defer server.Stop()

	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		log.Fatalf("Couldn't setup context, error: %s", err)
	}

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Rest.Host, c.Rest.Port)
	server.Start()
}
