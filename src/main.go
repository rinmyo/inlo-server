package main

import (
	"pracserver/src/interlock"
	"time"
)

func main() {
	if btn, ok := interlock.GetRouteByBtn("XLA", "S6LA"); ok {
		if ok := btn.Found(); ok {
			if btn, ok := interlock.GetRouteByName("ddd"); ok {
				if ok := btn.Found(); ok {
					timer := time.NewTimer(5 * time.Second)
					<-timer.C
					btn.Cancel()
				}
			}
		}
	}
}
