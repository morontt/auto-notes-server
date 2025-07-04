package models

import (
	"database/sql"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	pb "xelbot.com/auto-notes/server/rpc/server"
)

type FillingStation struct {
	ID        uint
	Name      string
	CreatedAt time.Time
}

type Fuel struct {
	ID        uint
	Cost      Cost
	Value     int32
	Station   FillingStation
	Date      time.Time
	Distance  sql.NullInt32
	Car       Car
	CreatedAt time.Time
}

func (f *Fuel) ToRpcMessage() *pb.Fuel {
	fuel := &pb.Fuel{
		Id: int32(f.ID),
		Car: &pb.Car{
			Id:   int32(f.Car.ID),
			Name: f.Car.Brand + " " + f.Car.Model,
		},
		Cost: &pb.Cost{
			Value:    f.Cost.Value,
			Currency: f.Cost.CurrencyCode,
		},
		Value: f.Value,
		Station: &pb.FillingStation{
			Id:        int32(f.Station.ID),
			Name:      f.Station.Name,
			CreatedAt: timestamppb.New(f.Station.CreatedAt),
		},
		Date:      timestamppb.New(f.Date),
		CreatedAt: timestamppb.New(f.CreatedAt),
	}

	if f.Distance.Valid {
		fuel.Distance = f.Distance.Int32
	}

	return fuel
}
