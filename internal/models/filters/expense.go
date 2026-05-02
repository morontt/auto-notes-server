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

func (p *ExpenseFilter) HasType() bool {
	if p == nil {
		return false
	}

	expType := p.pbFilter.GetType()

	return expType != pb.ExpenseType_EMPTY
}

func (p *ExpenseFilter) GetType() pb.ExpenseType {
	if p != nil {
		return p.pbFilter.GetType()
	}

	return pb.ExpenseType_EMPTY
}
