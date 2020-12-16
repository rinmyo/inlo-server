package main

import (
	log "github.com/sirupsen/logrus"
)

type SectionState byte

const (
	SectionBroken SectionState = iota
	SectionOccupied
	SectionFree
)

type ShuntingSignalState byte

const (
	ShuntingSignalL ShuntingSignalState = iota
	ShuntingSignalH
	ShuntingSignalU
	ShuntingSignalUU
	ShuntingSignalUUS
)

type StartingSignalState byte

const (
	StartingSignalL StartingSignalState = iota
	StartingSignalLU
	StartingSignalU
	StartingSignalH
	StartingSignalLL
)

type RouteSignalState byte

const (
	RouteSignalA RouteSignalState = iota
	RouteSignalB
	RouteSignalHB
)

type TurnoutState byte

const (
	TurnoutNormal TurnoutState = iota
	TurnoutReversed
	TurnoutBroken
)

func (turnoutState TurnoutState) String() string {
	switch turnoutState {
	case TurnoutNormal:
		return "Normal"
	case TurnoutReversed:
		return "Reversed"
	case TurnoutBroken:
		return "Broken"
	}
	return ""
}

func ReadSectionState(_ string) SectionState {
	// IO 接口電路
	return SectionFree
}

type Turnout struct {
	Tid   string
	State TurnoutState
}

func UpdateTurnoutState(turnout *Turnout, channel chan bool) {
	log.WithField("Turnout", turnout.Tid).WithField("State", turnout.State.String()).Info("update Turnout State")
	channel <- true
}

func SimulateReadTurnoutState(_ string, state TurnoutState) *TurnoutState {
	return &state
}
