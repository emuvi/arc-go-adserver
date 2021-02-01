package main

import (
	"log"
	"runtime"

	"adserver/biz"
	"adserver/guide"
	"adserver/motor"
)

func startCommandLine() {
	guide.Commands.Add(guide.Command("Production", true, "-p", "--production"))
	guide.Commands.Add(guide.Command("Speed", false, "-s", "--speed").DefaultValue("8"))
	guide.Commands.Add(guide.Command("MotorPort", false, "-mp", "--motor-port").DefaultValue("80"))
	guide.Commands.Add(guide.Command("StoreHost", false, "-sh", "--store-host").DefaultValue("pointel.pointto.us"))
	guide.Commands.Add(guide.Command("StorePort", false, "-sp", "--store-port").DefaultValue("5432"))
	guide.Commands.Parse().PutOnConfigs()
	runtime.GOMAXPROCS(guide.Configs.GetInt("Speed", 8))
}

func main() {
	startCommandLine()
	port := guide.Configs.GetInt("MotorPort", 80)
	log.Println("Starting AdServer at port", port, "...")
	biz.StartHandlers()
	motor.StartListen(port)
}
