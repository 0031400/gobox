package app

import (
	"gobox/config"
	"gobox/inbounds"
	"gobox/outbounds"
	"gobox/router"
	"log"
)

type App struct {
	filePath  string
	outbounds map[string]outbounds.Outbound
	inbounds  map[string]inbounds.Inbound
	router    *router.Router
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
	err = builder.LoadRuleSet()
	if err != nil {
		log.Println(err)
		return
	}
	dnsCenter, err := builder.BuildDnsCenter()
	if err != nil {
		log.Println(err)
		return
	}
	a.inbounds, err = builder.BuildInbounds()
	if err != nil {
		log.Println(err)
		return
	}
	a.outbounds, err = builder.BuildOutbounds()
	if err != nil {
		log.Println(err)
		return
	}
	a.router, err = builder.BuildRouter()
	if err != nil {
		log.Println(err)
		return
	}
	dnsCenter.Start()
	for tag, inbound := range a.inbounds {
		err = inbound.Start()
		if err != nil {
			log.Println(err)
			return
		}
		go a.handleInbound(tag, inbound)
	}
	select {}
}

func (a *App) handleInbound(tag string, inbound inbounds.Inbound) {
	for {
		session := inbound.Accept()
		log.Printf("[in] -> %s\n", tag)
		go func(inSession inbounds.InSession) {
			outboundTag := a.router.Route(session.Target)
			outbound, ok := a.outbounds[outboundTag]
			if !ok {
				log.Println("not found outbound")
				return
			}
			log.Printf("[route] %s -> %s\n", inSession.Target.ToString(), outboundTag)
			conn1 := inSession.Conn
			conn2, err := outbound.Connect(outbounds.OutSession{Target: inSession.Target, FirstData: inSession.FirstData})
			if err != nil {
				log.Println(err)
				return
			}
			finished := make(chan struct{}, 2)
			go func() {
				for {
					data, err := conn1.Read(4096)
					if err != nil {
						log.Println(err)
						break
					}
					err = conn2.Write(data)
					if err != nil {
						log.Println(err)
						break
					}
				}
				finished <- struct{}{}
			}()
			go func() {
				for {
					data, err := conn2.Read(4096)
					if err != nil {
						log.Println(err)
						break
					}
					err = conn1.Write(data)
					if err != nil {
						log.Println(err)
						break
					}
				}
				finished <- struct{}{}
			}()
			<-finished
			<-finished
		}(session)
	}
}
