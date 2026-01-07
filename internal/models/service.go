package models

import (
	"database/sql"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	pb "xelbot.com/auto-notes/server/rpc/server"
)

type Service struct {
	ID          uint
	Cost        *Cost
	Description string
	Date        time.Time
	Distance    sql.NullInt32
	Mileage     *Mileage
	Car         *Car
	CreatedAt   time.Time
}

func (s *Service) ToRpcMessage() *pb.Service {
	message := &pb.Service{
		Id:          int32(s.ID),
		Description: s.Description,
		Date:        timestamppb.New(s.Date),
		CreatedAt:   timestamppb.New(s.CreatedAt),
	}

	if s.Car != nil {
		message.Car = &pb.Car{
			Id:   int32(s.Car.ID),
			Name: s.Car.Brand + " " + s.Car.Model,
		}
	}

	if s.Cost != nil {
		message.Cost = &pb.Cost{
			Value:    s.Cost.Value,
			Currency: s.Cost.CurrencyCode,
		}
	}

	if s.Distance.Valid {
		message.Distance = s.Distance.Int32
	}

	return message
}
