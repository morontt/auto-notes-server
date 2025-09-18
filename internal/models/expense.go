package models

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	pb "xelbot.com/auto-notes/server/rpc/server"
)

type Expense struct {
	ID          uint
	Cost        Cost
	Description string
	Date        time.Time
	Car         *Car
	Type        pb.ExpenseType
	CreatedAt   time.Time
}

func (e *Expense) ToRpcMessage() *pb.Expense {
	message := &pb.Expense{
		Id: int32(e.ID),
		Cost: &pb.Cost{
			Value:    e.Cost.Value,
			Currency: e.Cost.CurrencyCode,
		},
		Description: e.Description,
		Type:        e.Type,
		Date:        timestamppb.New(e.Date),
		CreatedAt:   timestamppb.New(e.CreatedAt),
	}

	if e.Car != nil {
		message.Car = &pb.Car{
			Id:   int32(e.Car.ID),
			Name: e.Car.Brand + " " + e.Car.Model,
		}
	}

	return message
}
