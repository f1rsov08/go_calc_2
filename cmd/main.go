package main

import (
	"github.com/f1rsov08/go_calc_2/internal/agent"
	"github.com/f1rsov08/go_calc_2/internal/orchestrator"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	orchestrator_app := orchestrator.New()
	agent_app := agent.New()
	go orchestrator_app.RunServer()
	go agent_app.Run()
	select {}
}
