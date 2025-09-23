package server

import (
	"context"
	"database/sql"
	"errors"

	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/types/known/emptypb"
	"xelbot.com/auto-notes/server/internal/application"
	"xelbot.com/auto-notes/server/internal/models"
	"xelbot.com/auto-notes/server/internal/models/filters"
	"xelbot.com/auto-notes/server/internal/models/repository"
	pb "xelbot.com/auto-notes/server/rpc/server"
)

type OrderRepositoryService struct {
	app application.Container
}

func NewOrderRepositoryService(app application.Container) *OrderRepositoryService {
	return &OrderRepositoryService{app: app}
}

func (or *OrderRepositoryService) GetOrders(ctx context.Context, pbFilter *pb.OrderFilter) (*pb.OrderCollection, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	filter := filters.NewOrderFilter(pbFilter)

	repo := repository.OrderRepository{DB: or.app.DB}
	dbOrders, cntOrders, err := repo.GetOrdersByUser(user.ID, filter)
	if err != nil {
		or.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	if pageOutOfRange(filter, cntOrders) {
		return nil, twirp.NotFoundError("orders not found")
	}

	orders := make([]*pb.Order, 0, len(dbOrders))
	for _, dbOrder := range dbOrders {
		orders = append(orders, dbOrder.ToRpcMessage())
	}

	or.app.Info("OrderRepositoryService: populate orders", ctx, "cnt", len(dbOrders))

	return &pb.OrderCollection{
		Orders: orders,
		Meta: &pb.PaginationMeta{
			Current: int32(filter.GetPage()),
			Last:    int32(filters.GetLastPage(filter, cntOrders)),
		},
	}, nil
}

func (or *OrderRepositoryService) FindOrder(ctx context.Context, idReq *pb.IdRequest) (*pb.Order, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	repo := repository.OrderRepository{DB: or.app.DB}
	if idReq.GetId() > 0 {
		ownerId, err := repo.OrderOwner(uint(idReq.GetId()))
		if err != nil {
			if errors.Is(err, models.RecordNotFound) {
				return nil, twirp.NotFound.Error("order not found")
			} else {
				or.app.ServerError(ctx, err)

				return nil, twirp.InternalError("internal error")
			}
		}
		if ownerId != user.ID {
			return nil, twirp.InvalidArgument.Error("invalid order owner")
		}
	} else {
		return nil, twirp.InvalidArgument.Error("invalid id")
	}

	dbOrder, err := repo.Find(uint(idReq.GetId()))
	if err != nil {
		or.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	return dbOrder.ToRpcMessage(), nil
}

func (or *OrderRepositoryService) GetOrderTypes(ctx context.Context, _ *emptypb.Empty) (*pb.OrderTypeCollection, error) {
	_, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	repo := repository.OrderRepository{DB: or.app.DB}
	dbTypes, err := repo.GetOrderTypes()
	if err != nil {
		or.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	types := make([]*pb.OrderType, 0, len(dbTypes))
	for _, dbItem := range dbTypes {
		item := &pb.OrderType{
			Id:   int32(dbItem.ID),
			Name: dbItem.Name,
		}

		types = append(types, item)
	}

	or.app.Info("OrderRepositoryService: populate order types", ctx, "cnt", len(dbTypes))

	return &pb.OrderTypeCollection{Types: types}, nil
}

func (or *OrderRepositoryService) SaveOrder(ctx context.Context, order *pb.Order) (*pb.Order, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	currencyCode := order.Cost.GetCurrency()
	if currencyCode == "" {
		return nil, twirp.InvalidArgument.Error("empty currency code")
	}

	if order.GetDate() == nil {
		return nil, twirp.InvalidArgument.Error("date is required")
	}

	orderRepo := repository.OrderRepository{DB: or.app.DB}
	if order.GetId() > 0 {
		ownerId, err := orderRepo.OrderOwner(uint(order.GetId()))
		if err != nil {
			if errors.Is(err, models.RecordNotFound) {
				return nil, twirp.NotFound.Error("order not found")
			} else {
				or.app.ServerError(ctx, err)

				return nil, twirp.InternalError("internal error")
			}
		}
		if ownerId != user.ID {
			return nil, twirp.InvalidArgument.Error("invalid order owner")
		}
	}

	var orderType *models.OrderType
	if order.Type.GetId() > 0 {
		orderType, err = orderRepo.FindType(uint(order.Type.GetId()))
		if err != nil {
			if errors.Is(err, models.RecordNotFound) {
				return nil, twirp.InvalidArgument.Error("invalid order type")
			} else {
				or.app.ServerError(ctx, err)

				return nil, twirp.InternalError("internal error")
			}
		}
	}

	currencyRepo := repository.CurrencyRepository{DB: or.app.DB}
	currency, err := currencyRepo.GetCurrencyByCode(currencyCode)
	if err != nil {
		if errors.Is(err, models.RecordNotFound) {
			return nil, twirp.InvalidArgument.Error("invalid currency")
		} else {
			or.app.ServerError(ctx, err)

			return nil, twirp.InternalError("internal error")
		}
	}

	var car *models.Car
	if order.Car.GetId() > 0 {
		carRepo := repository.CarRepository{DB: or.app.DB}
		car, err = carRepo.Find(uint(order.Car.GetId()))
		if err != nil {
			if errors.Is(err, models.RecordNotFound) {
				return nil, twirp.InvalidArgument.Error("invalid car")
			} else {
				or.app.ServerError(ctx, err)

				return nil, twirp.InternalError("internal error")
			}
		}

		if car.UserID != user.ID {
			return nil, twirp.InvalidArgument.Error("invalid car owner")
		}
	}

	/*
		Distance    sql.NullInt32
	*/

	var capacity sql.NullString
	if order.GetCapacity() == "" {
		capacity = sql.NullString{Valid: false}
	} else {
		capacity = sql.NullString{
			String: order.GetCapacity(),
			Valid:  true,
		}
	}

	var usedAt sql.NullTime
	if order.GetUsedAt() != nil {
		usedAt = sql.NullTime{
			Valid: true,
			Time:  order.GetUsedAt().AsTime(),
		}
	} else {
		usedAt = sql.NullTime{Valid: false}
	}

	orderModel := models.Order{
		ID:  uint(order.GetId()),
		Car: car,
		Cost: models.Cost{
			Value:      order.Cost.GetValue(),
			CurrencyID: currency.ID,
		},
		Description: order.GetDescription(),
		Capacity:    capacity,
		Date:        order.Date.AsTime(),
		Type:        orderType,
		UsedAt:      usedAt,
	}

	orderID, err := orderRepo.SaveOrder(&orderModel, user.ID)
	if err != nil {
		or.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	dbOrder, err := orderRepo.Find(orderID)
	if err != nil {
		or.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	return dbOrder.ToRpcMessage(), nil
}

func (or *OrderRepositoryService) GetExpenses(ctx context.Context, pbFilter *pb.ExpenseFilter) (*pb.ExpenseCollection, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	filter := filters.NewExpenseFilter(pbFilter)

	repo := repository.ExpenseRepository{DB: or.app.DB}
	dbExpenses, cntExpenses, err := repo.GetExpensesByUser(user.ID, filter)
	if err != nil {
		or.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	if pageOutOfRange(filter, cntExpenses) {
		return nil, twirp.NotFoundError("expenses not found")
	}

	expenses := make([]*pb.Expense, 0, len(dbExpenses))
	for _, dbExpense := range dbExpenses {
		expenses = append(expenses, dbExpense.ToRpcMessage())
	}

	or.app.Info("OrderRepositoryService: populate expenses", ctx, "cnt", len(dbExpenses))

	return &pb.ExpenseCollection{
		Expenses: expenses,
		Meta: &pb.PaginationMeta{
			Current: int32(filter.GetPage()),
			Last:    int32(filters.GetLastPage(filter, cntExpenses)),
		},
	}, nil
}

func (or *OrderRepositoryService) FindExpense(ctx context.Context, idReq *pb.IdRequest) (*pb.Expense, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	repo := repository.ExpenseRepository{DB: or.app.DB}
	if idReq.GetId() > 0 {
		ownerId, err := repo.ExpenseOwner(uint(idReq.GetId()))
		if err != nil {
			if errors.Is(err, models.RecordNotFound) {
				return nil, twirp.NotFound.Error("expense not found")
			} else {
				or.app.ServerError(ctx, err)

				return nil, twirp.InternalError("internal error")
			}
		}
		if ownerId != user.ID {
			return nil, twirp.InvalidArgument.Error("invalid expense owner")
		}
	} else {
		return nil, twirp.InvalidArgument.Error("invalid id")
	}

	dbExpense, err := repo.Find(uint(idReq.GetId()))
	if err != nil {
		or.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	return dbExpense.ToRpcMessage(), nil
}

func (or *OrderRepositoryService) SaveExpense(ctx context.Context, expense *pb.Expense) (*pb.Expense, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	currencyCode := expense.Cost.GetCurrency()
	if currencyCode == "" {
		return nil, twirp.InvalidArgument.Error("empty currency code")
	}

	if expense.GetType() == pb.ExpenseType_EMPTY {
		return nil, twirp.InvalidArgument.Error("expense type is required")
	}

	if expense.GetDate() == nil {
		return nil, twirp.InvalidArgument.Error("date is required")
	}

	expenseRepo := repository.ExpenseRepository{DB: or.app.DB}
	if expense.GetId() > 0 {
		ownerId, err := expenseRepo.ExpenseOwner(uint(expense.GetId()))
		if err != nil {
			if errors.Is(err, models.RecordNotFound) {
				return nil, twirp.NotFound.Error("order not found")
			} else {
				or.app.ServerError(ctx, err)

				return nil, twirp.InternalError("internal error")
			}
		}
		if ownerId != user.ID {
			return nil, twirp.InvalidArgument.Error("invalid expense owner")
		}
	}

	currencyRepo := repository.CurrencyRepository{DB: or.app.DB}
	currency, err := currencyRepo.GetCurrencyByCode(currencyCode)
	if err != nil {
		if errors.Is(err, models.RecordNotFound) {
			return nil, twirp.InvalidArgument.Error("invalid currency")
		} else {
			or.app.ServerError(ctx, err)

			return nil, twirp.InternalError("internal error")
		}
	}

	var car *models.Car
	if expense.Car.GetId() > 0 {
		carRepo := repository.CarRepository{DB: or.app.DB}
		car, err = carRepo.Find(uint(expense.Car.GetId()))
		if err != nil {
			if errors.Is(err, models.RecordNotFound) {
				return nil, twirp.InvalidArgument.Error("invalid car")
			} else {
				or.app.ServerError(ctx, err)

				return nil, twirp.InternalError("internal error")
			}
		}

		if car.UserID != user.ID {
			return nil, twirp.InvalidArgument.Error("invalid car owner")
		}
	}

	expenseModel := models.Expense{
		ID:  uint(expense.GetId()),
		Car: car,
		Cost: models.Cost{
			Value:      expense.GetCost().GetValue(),
			CurrencyID: currency.ID,
		},
		Description: expense.GetDescription(),
		Date:        expense.GetDate().AsTime(),
		Type:        expense.GetType(),
	}

	expenseID, err := expenseRepo.SaveExpense(&expenseModel, user.ID)
	if err != nil {
		or.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	dbExpense, err := expenseRepo.Find(expenseID)
	if err != nil {
		or.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	return dbExpense.ToRpcMessage(), nil
}
