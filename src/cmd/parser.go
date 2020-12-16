package main

import (
	"errors"
	"regexp"
	"strings"
)

func ParseTurnout(turnout string) ([]*Turnout, error) {
	//ss
	if matched, _ := regexp.MatchString("^[0-9]*$", turnout); matched {
		return []*Turnout{{
			Tid:   turnout,
			State: TurnoutNormal,
		}}, nil
	}

	//(ss)
	if matched, _ := regexp.MatchString("^\\([0-9]*\\)$", turnout); matched {
		return []*Turnout{{
			Tid:   turnout,
			State: TurnoutReversed,
		}}, nil
	}

	//ss/ss
	if matched, _ := regexp.MatchString("^[0-9]*/[0-9]*$", turnout); matched {
		slashPosition := strings.Index(turnout, "/")
		return []*Turnout{
			{
				Tid:   turnout[:slashPosition],
				State: TurnoutNormal,
			},
			{
				Tid:   turnout[slashPosition+1:],
				State: TurnoutNormal,
			},
		}, nil
	}

	//(ss/ss)
	if matched, _ := regexp.MatchString("^\\([0-9]*/[0-9]*\\)$", turnout); matched {
		slashPosition := strings.Index(turnout, "/")
		return []*Turnout{
			{
				Tid:   turnout[1:slashPosition],
				State: TurnoutReversed,
			},
			{
				Tid:   turnout[slashPosition+1 : len(turnout)-1],
				State: TurnoutReversed,
			},
		}, nil
	}

	return nil, errors.New("cannot match any pattern")
}
