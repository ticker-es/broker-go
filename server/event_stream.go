package server

import (
	"context"
	"errors"

	"github.com/golang/protobuf/ptypes"

	es "github.com/ticker-es/client-go/eventstream/base"

	"github.com/ticker-es/client-go/rpc"
)

type eventStreamServer struct {
	rpc.UnimplementedEventStreamServer
	server *Server
}

func (s *eventStreamServer) Emit(ctx context.Context, event *rpc.Event) (*rpc.Ack, error) {
	occurredAt, err := ptypes.Timestamp(event.OccurredAt)
	if err != nil {
		return nil, err
	}
	payload := event.Payload.AsMap()
	ev := es.Event{
		Sequence:   -1,
		Aggregate:  event.Aggregate,
		Type:       event.Type,
		OccurredAt: occurredAt,
		Payload:    payload,
	}
	seq, err := s.server.streamBackend.Store(&ev)
	return &rpc.Ack{
		Sequence: int64(seq),
	}, err
}

func (s *eventStreamServer) Stream(req *rpc.StreamRequest, stream rpc.EventStream_StreamServer) error {
	selector := es.Selector{
		Aggregate: req.Selector.Aggregate,
		Type:      req.Selector.Type,
	}
	bracket := es.Bracket{
		NextSequence: req.Bracket.FirstSequence,
		LastSequence: req.Bracket.LastSequence,
	}
	return s.server.streamBackend.Stream(stream.Context(), selector, bracket, func(e *es.Event) error {
		ev := rpc.EventToProto(e)
		return stream.Send(ev)
	})
}

func (s *eventStreamServer) Subscribe(req *rpc.SubscriptionRequest, stream rpc.EventStream_SubscribeServer) error {
	if req == nil {
		return nil
	}
	persistentClientID := req.PersistentClientId
	if req.Selector == nil {
		return errors.New("called with nil selector")
	}
	selector := es.Selector{
		Aggregate: req.Selector.Aggregate,
		Type:      req.Selector.Type,
	}
	sub, err := s.server.streamBackend.Subscribe(stream.Context(), persistentClientID, selector, func(e *es.Event) error {
		ev := rpc.EventToProto(e)
		return stream.Send(ev)
	})
	if err != nil {
		return err
	}
	return sub.Wait()
}
