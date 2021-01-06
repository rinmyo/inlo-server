package service

import (
	"context"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"pracserver/src/config"
	"pracserver/src/pb"
	"reflect"
)

// StationServer is the server that provides station service
type StationServer struct {
	sm *StationManager
}

func (s *StationServer) ManualUnlockRoute(ctx context.Context, request *pb.ManualUnlockRouteRequest) (*emptypb.Empty, error) {
	panic("implement me")
}

func (s *StationServer) ErrorUnlockRoute(ctx context.Context, request *pb.ErrorUnlockRouteRequest) (*emptypb.Empty, error) {
	panic("implement me")
}

// NewStationServer returns a new StationServer object
func NewStationServer(statusManager *StationManager) *StationServer {
	return &StationServer{statusManager}
}

func (s *StationServer) InitStation(context.Context, *emptypb.Empty) (*pb.InitStationResponse, error) {
	var signals []*pb.Signal
	for id, state := range s.sm.signals {
		signals = append(signals, &pb.Signal{
			Id:    id,
			State: state,
		})
	}

	var turnouts []*pb.Turnout
	for id, state := range s.sm.turnouts {
		turnouts = append(turnouts, &pb.Turnout{
			Id:    id,
			State: state,
		})
	}

	var sections []*pb.Section
	for id, state := range s.sm.sections {
		sections = append(sections, &pb.Section{
			Id:    id,
			State: state,
		})
	}

	var routes []*pb.InitRouteMessage
	for _, r := range s.sm.AliveRoutes() {
		routes = append(routes,
			&pb.InitRouteMessage{
				RouteId:  r.Id,
				ButtonId: r.Buttons[0],
			})
	}

	response := &pb.InitStationResponse{
		Signal:  signals,
		Turnout: turnouts,
		Section: sections,
		Route:   routes,
	}
	return response, nil
}

func (s *StationServer) RefreshStation(_ *emptypb.Empty, stream pb.StationService_RefreshStationServer) error {
	close(s.sm.channel) //關閉舊信道
	ioInfo := (*s.sm.controller).GetIOInfo()
	s.sm.channel = make(chan *StateChangedEvent, len(ioInfo["turnouts"])+len(ioInfo["sections"])+len(ioInfo["signals"]))

	for e := range s.sm.channel {
		response := &pb.RefreshStationResponse{}
		switch e.st {
		case SectionType:
			response.Value = &pb.RefreshStationResponse_Section{
				Section: &pb.Section{Id: e.id, State: pb.Section_SectionState(e.new)},
			}
		case TurnoutType:
			response.Value = &pb.RefreshStationResponse_Turnout{
				Turnout: &pb.Turnout{Id: e.id, State: pb.Turnout_TurnoutState(e.new)},
			}
		case SignalType:
			response.Value = &pb.RefreshStationResponse_Signal{
				Signal: &pb.Signal{Id: e.id, State: pb.Signal_SignalState(e.new)},
			}
		}
		err := stream.Send(response)
		if err != nil {
			log.Error(err)
			return err
		}
		log.Info("sent a station: ", response)
	}
	log.Info("一次會話結束")
	return nil
}

func (s *StationServer) getRouteByName(name string) (*Route, bool) {
	val, ok := s.sm.interlock[name]
	if !ok {
		log.WithField(reason, config.Msg.NoSuchRouteMsg).
			WithField(routeName, name).
			Error(config.Msg.ObtainRouteFailMsg)
	}
	return val, ok
}

func (s *StationServer) getRouteByBtn(btns ...string) (*Route, bool) {
	for _, v := range s.sm.interlock {
		if reflect.DeepEqual(v.Buttons, btns) {
			return v, true
		}
	}
	log.WithField(reason, config.Msg.NoSuchRouteMsg).
		WithField(buttons, btns).
		Error(config.Msg.ObtainRouteFailMsg)
	return nil, false
}

func (s *StationServer) CreateRoute(_ context.Context, req *pb.CreateRouteRequest) (*pb.CreateRouteResponse, error) {
	btns := req.GetButtons().GetButtonId()
	log.Debug("get buttons: ", btns)
	if route, ok := s.getRouteByBtn(btns...); ok {
		err := s.sm.CreateRoute(route)
		if err != nil {
			return nil, err
		}
		return &pb.CreateRouteResponse{
			RouteId: route.Id,
		}, nil
	}
	return nil, status.Error(codes.NotFound, "not found the route")
}

func (s *StationServer) CancelRoute(_ context.Context, req *pb.CancelRouteRequest) (*emptypb.Empty, error) {
	rid := req.GetRouteId()
	route, ok := s.getRouteByName(rid)
	if ok {
		err := s.sm.CancelRoute(route)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		log.Info("create route command: " + rid)
		return &emptypb.Empty{}, nil
	}
	return nil, status.Errorf(codes.NotFound, "no such route: %s", rid)
}
