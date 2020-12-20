package service

import (
	"errors"
	"pracserver/src/pb"
	"regexp"
	"strings"
)

func ParseTurnout(str string) ([]*pb.Turnout, error) {
	//ss
	if matched, _ := regexp.MatchString("^[0-9]*$", str); matched {
		return []*pb.Turnout{{
			Id:    str,
			State: pb.Turnout_NORMAL,
		}}, nil
	}

	//(ss)
	if matched, _ := regexp.MatchString("^\\([0-9]*\\)$", str); matched {
		return []*pb.Turnout{{
			Id:    str,
			State: pb.Turnout_REVERSED,
		}}, nil
	}

	//ss/ss
	if matched, _ := regexp.MatchString("^[0-9]*/[0-9]*$", str); matched {
		slashPosition := strings.Index(str, "/")
		return []*pb.Turnout{
			{
				Id:    str[:slashPosition],
				State: pb.Turnout_NORMAL,
			},
			{
				Id:    str[slashPosition+1:],
				State: pb.Turnout_NORMAL,
			},
		}, nil
	}

	//(ss/ss)
	if matched, _ := regexp.MatchString("^\\([0-9]*/[0-9]*\\)$", str); matched {
		slashPosition := strings.Index(str, "/")
		return []*pb.Turnout{
			{
				Id:    str[1:slashPosition],
				State: pb.Turnout_REVERSED,
			},
			{
				Id:    str[slashPosition+1 : len(str)-1],
				State: pb.Turnout_REVERSED,
			},
		}, nil
	}

	return nil, errors.New("cannot match any pattern")
}

func ParseOccupiedSection(str string) *pb.Section {
	return &pb.Section{
		Id:    str,
		State: pb.Section_OCCUPIED,
	}
}

func ParseSignal(str string) *pb.Signal {
	p := strings.Index(str, "-")
	id := str[:p]
	var state pb.Signal_SignalState
	switch str[p+1:] {
	case "H":
		state = pb.Signal_RED
	case "U":
		state = pb.Signal_YELLOW
	case "UU":
		state = pb.Signal_DOUBLE_YELLOW
	case "L":
		state = pb.Signal_GREEN
	case "B":
		state = pb.Signal_WHITE
	}
	return &pb.Signal{
		Id:    id,
		State: state,
	}
}
