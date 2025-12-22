package models

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	pb "xelbot.com/auto-notes/server/rpc/server"
)

type Mileage struct {
	ID        uint
	Distance  int
	Date      time.Time
	Car       *Car
	CreatedAt time.Time
}

func (m *Mileage) ToRpcMessage() *pb.Mileage {
	message := &pb.Mileage{
		Id:        int32(m.ID),
		Distance:  int32(m.Distance),
		Date:      timestamppb.New(m.Date),
		CreatedAt: timestamppb.New(m.CreatedAt),
	}

	if m.Car != nil {
		message.Car = &pb.Car{
			Id:   int32(m.Car.ID),
			Name: m.Car.Brand + " " + m.Car.Model,
		}
	}

	return message
}
