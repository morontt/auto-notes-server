package filters

import pb "xelbot.com/auto-notes/server/rpc/server"

type ExpenseFilter struct {
	pbFilter *pb.ExpenseFilter
	pager
}

func NewExpenseFilter(f *pb.ExpenseFilter) *ExpenseFilter {
	return &ExpenseFilter{
		pbFilter: f,
		pager:    pager{filter: f},
	}
}
