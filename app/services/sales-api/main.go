package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/iBoBoTi/service-ardan/foundation/logger"
	"github.com/iBoBoTi/service-ardan/business/web/v1/debug"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

var build = "develop"

func main(){
	log, err := logger.New("SALES-API")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	if err := run(log); err != nil {
		log.Errorw("start up", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error{
	// ===========================================================================================
	// GOMAXPROCS
	opts := maxprocs.Logger(log.Infof)
	if _, err := maxprocs.Set(opts); err != nil {
		return fmt.Errorf("maxprocs: %w", err)
	}
	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))
	defer log.Infow("shutdown")

	// ===========================================================================================
	// Configuration
	cfg := struct{
		conf.Version
		Web struct{
			ReadTimeout time.Duration `conf:"default:5s"`
			WriteTimeout time.Duration `conf:"default:5s"`
			IdleTimeout time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s,mask"` //mask or noprint
			APIHost string `conf:"default:0.0.0.0:3000"`
			DebugHost string `conf:"default:0.0.0.0:4000"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc: "copyright information here",
		},
	}

	const prefix = "SALES" 
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted){
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// ===========================================================================================
	// App Starting
	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("startup", "config", out)

	// ===========================================================================================
	// Start Debug Service
	log.Infow("startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)
	go func(){
		if err := http.ListenAndServe(cfg.Web.DebugHost, debug.StandardLibraryMux()); err != nil {
			log.Errorw("shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "ERROR", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	return nil
}