package main

import (
	"fmt"
	"os"

	"github.com/iBoBoTi/service-ardan/foundation/logger"
	"go.uber.org/zap"
)

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
	return nil
}