package app

import (
	"pracserver/internal/app/service"
)

type App struct {
	im *service.InstanceManager
	sm *service.StationManager
}

func NewApp() *App {
	return &App{}
}

func NewAppServer(app *App) {

}
