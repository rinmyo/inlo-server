package main

import (
	log "github.com/sirupsen/logrus"
	"pracserver/src/interlock"
	"time"
)

func main() {
	if btn, err := interlock.GetRouteByBtn("XLA", "S6LA"); err != nil {
		log.Error(err)
	} else {
		if ok := btn.Found(); ok {
			if btn, err := interlock.GetRouteByBtn("XLA", "S6LA"); err != nil {
				log.Error(err)
			} else {
				if ok := btn.Found(); ok {
					timer := time.NewTimer(5 * time.Second)
					<-timer.C
					btn.Cancel()
				}
			}
		}
	}
}
