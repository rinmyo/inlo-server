package main

import log "github.com/sirupsen/logrus"

func main() {
	if btn, err := getRouteByBtn("XLA", "S6LA"); err != nil {
		log.Error(err)
	} else {
		log.Info(btn)

		log.Info(btn.isLiving())
		if err := btn.found(); err != nil {
			log.Error(err)
		} else {
			log.Info("new living sinro", btn.Id)
		}

		log.Info(btn.isLiving())
		if err := btn.found(); err != nil {
			log.Error(err)
		} else {
			log.Info("new living sinro", btn.Id)
		}

	}
}
