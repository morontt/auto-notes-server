package filters

import pb "xelbot.com/auto-notes/server/rpc/server"

type ExpenseFilter struct {
	pbFilter *pb.ExpenseFilter
}

func NewExpenseFilter(f *pb.ExpenseFilter) *ExpenseFilter {
	return &ExpenseFilter{pbFilter: f}
}

func (ef *ExpenseFilter) GetPage() int {
	if ef.pbFilter.GetPage() > 0 {
		return int(ef.pbFilter.GetPage())
	}

	return 1
}

func (ef *ExpenseFilter) GetLimit() int {
	return int(ef.pbFilter.GetLimit())
}
