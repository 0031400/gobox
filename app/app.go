package app

import (
	"gobox/config"
	"log"
)

type App struct {
	filePath string
}

func NewApp(filePath string) *App {
	return &App{filePath: filePath}
}
func (a *App) Run() {
	appConfig, err := config.LoadConfig(a.filePath)
	if err != nil {
		log.Println(err)
		return
	}
	builder := NewBuilder(*appConfig)
	dnsCenter, err := builder.BuildDnsCenter()
	if err != nil {
		log.Println(err)
		return
	}
	dnsCenter.Start()
	channel := make(chan struct{})
	<-channel
}
