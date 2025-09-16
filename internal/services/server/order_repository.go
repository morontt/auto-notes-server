package server

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"xelbot.com/auto-notes/server/internal/application"
	pb "xelbot.com/auto-notes/server/rpc/server"
)

type OrderRepositoryService struct {
	app application.Container
}

func NewOrderRepositoryService(app application.Container) *OrderRepositoryService {
	return &OrderRepositoryService{app: app}
}

func (or *OrderRepositoryService) GetOrders(ctx context.Context, filter *pb.OrderFilter) (*pb.OrderCollection, error) {
	return nil, nil
}

func (or *OrderRepositoryService) GetOrderTypes(ctx context.Context, _ *emptypb.Empty) (*pb.OrderTypeCollection, error) {
	return nil, nil
}

func (or *OrderRepositoryService) SaveOrder(ctx context.Context, order *pb.Order) (*pb.Order, error) {
	return nil, nil
}

func (or *OrderRepositoryService) GetExpenses(ctx context.Context, filter *pb.ExpenseFilter) (*pb.ExpenseCollection, error) {
	return nil, nil
}

func (or *OrderRepositoryService) SaveExpense(ctx context.Context, expense *pb.Expense) (*pb.Expense, error) {
	return nil, nil
}
