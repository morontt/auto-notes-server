package filters

import pb "xelbot.com/auto-notes/server/rpc/server"

type ExpenseFilter struct {
	pbFilter *pb.ExpenseFilter
	commonPart
}

func NewExpenseFilter(f *pb.ExpenseFilter) *ExpenseFilter {
	return &ExpenseFilter{
		pbFilter:   f,
		commonPart: commonPart{filter: f},
	}
}
