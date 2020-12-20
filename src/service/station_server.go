package service

import (
	"context"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"pracserver/src/pb"
)

// StationServer is the server that provides station service
type StationServer struct {
	sm *StationManager
}

// NewStationServer returns a new StationServer object
func NewStationServer(statusManager *StationManager) *StationServer {
	return &StationServer{statusManager}
}

func (s *StationServer) RefreshStation(_ *emptypb.Empty, stream pb.StationService_RefreshStationServer) error {
	s.sm.OnStatusChange(
		func(e StatusChangedEvent) error {
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
				return err
			}
			log.Info("sent a station")
			return nil
		})
	return nil
}

func (s *StationServer) SearchRoute(_ context.Context, req *pb.SearchRouteRequest) (*pb.SearchRouteResponse, error) {
	btns := req.GetButtons().GetButtonId()
	if route, ok := s.sm.GetRouteByBtn(btns...); ok {
		var (
			turnouts []*pb.Turnout
			signals  []*pb.Signal
			sections []*pb.Section
		)
		for _, v := range route.Turnouts {
			t, err := ParseTurnout(v)
			if err != nil {
				return nil, err
			}
			log.Debug("search turnout: ", v, "->", t)
			turnouts = append(turnouts, t...)
		}
		for _, v := range route.Signals {
			s := ParseSignal(v)
			signals = append(signals, s)
			log.Debug("search signal: ", v, "->", s)
		}
		for _, v := range route.Sections {
			s := ParseOccupiedSection(v)
			sections = append(sections, s)
			log.Debug("search section: ", v, "->", s)
		}

		return &pb.SearchRouteResponse{
			Route: &pb.Route{
				Id:       route.Id,
				Sections: sections,
				Turnouts: turnouts,
				Signals:  signals,
			},
		}, nil
	}
	return nil, status.Error(codes.NotFound, "not found the path")
}

func (s *StationServer) CreateRoute(_ context.Context, req *pb.NewRouteRequest) (*emptypb.Empty, error) {
	rid := req.GetRouteId()
	route, ok := s.sm.GetRouteByName(rid)
	if ok {
		err := s.sm.CreateRoute(route)
		if err != nil {
			return nil, err
		}
		log.Info("create route command: " + rid)
		return &emptypb.Empty{}, nil
	}
	return nil, status.Errorf(codes.NotFound, "no such route: %s", rid)
}

func (s *StationServer) CancelRoute(_ context.Context, req *pb.CancelRouteRequest) (*emptypb.Empty, error) {
	rid := req.GetRouteId()
	route, ok := s.sm.GetRouteByName(rid)
	if ok {
		s.sm.CancelRoute(route)
		log.Info("create route command: " + rid)
		return &emptypb.Empty{}, nil
	}
	return nil, status.Errorf(codes.NotFound, "no such route: %s", rid)
}
