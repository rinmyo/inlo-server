package service

import (
	"errors"
	"pracserver/src/pb"
	"regexp"
	"strings"
)

func ParseNormalTurnout(str string) ([]*pb.Turnout, error) {
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
			Id:    str[1 : len(str)-1],
			State: pb.Turnout_NORMAL,
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
				State: pb.Turnout_NORMAL,
			},
			{
				Id:    str[slashPosition+1 : len(str)-1],
				State: pb.Turnout_NORMAL,
			},
		}, nil
	}

	return nil, errors.New("cannot match any pattern: " + str)
}

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
			Id:    str[1 : len(str)-1],
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

	return nil, errors.New("cannot match any pattern: " + str)
}

func ParseSection(str string, state pb.Section_SectionState) *pb.Section {
	return &pb.Section{
		Id:    str,
		State: state,
	}
}

func ParseAbortSignal(str string) *pb.Signal {
	p := strings.Index(str, "-")
	id := str[:p]
	state := pb.Signal_UNKNOWN

	//調車信號機
	if matched, _ := regexp.MatchString("^D[0-9]+$", id); matched {
		state = pb.Signal_BLUE
	} else {
		state = pb.Signal_RED
	}
	return &pb.Signal{
		Id:    id,
		State: state,
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
