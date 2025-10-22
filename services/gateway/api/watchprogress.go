package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"watch-progress-service/services/gateway/api/internal/config"
	"watch-progress-service/services/gateway/api/internal/handler"
	"watch-progress-service/services/gateway/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", os.Getenv("WATCH_PROGRESS_CONFIG_FILE"), "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.Rest)
	defer server.Stop()

	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		log.Fatalf("Coldn't setup context, error: %s", err)
	}

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Rest.Host, c.Rest.Port)
	server.Start()
}
