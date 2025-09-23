package models

import (
	"database/sql"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	pb "xelbot.com/auto-notes/server/rpc/server"
)

type OrderType struct {
	ID   uint
	Name string
}

type Order struct {
	ID          uint
	Cost        Cost
	Description string
	Capacity    sql.NullString
	Date        time.Time
	UsedAt      sql.NullTime
	Distance    sql.NullInt32
	Car         *Car
	Type        *OrderType
	CreatedAt   time.Time
}

func (o *Order) ToRpcMessage() *pb.Order {
	message := &pb.Order{
		Id: int32(o.ID),
		Cost: &pb.Cost{
			Value:    o.Cost.Value,
			Currency: o.Cost.CurrencyCode,
		},
		Description: o.Description,
		Date:        timestamppb.New(o.Date),
		CreatedAt:   timestamppb.New(o.CreatedAt),
	}

	if o.Type != nil {
		message.Type = &pb.OrderType{
			Id:   int32(o.Type.ID),
			Name: o.Type.Name,
		}
	}

	if o.Car != nil {
		message.Car = &pb.Car{
			Id:   int32(o.Car.ID),
			Name: o.Car.Brand + " " + o.Car.Model,
		}
	}

	if o.UsedAt.Valid {
		message.UsedAt = timestamppb.New(o.UsedAt.Time)
	}

	if o.Capacity.Valid {
		message.Capacity = o.Capacity.String
	}

	if o.Distance.Valid {
		message.Distance = o.Distance.Int32
	}

	return message
}
