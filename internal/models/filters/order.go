package filters

import pb "xelbot.com/auto-notes/server/rpc/server"

type OrderFilter struct {
	pbFilter *pb.OrderFilter
}

func NewOrderFilter(f *pb.OrderFilter) *OrderFilter {
	return &OrderFilter{pbFilter: f}
}

func (of *OrderFilter) GetPage() int {
	if of.pbFilter.GetPage() > 0 {
		return int(of.pbFilter.GetPage())
	}

	return 1
}

func (of *OrderFilter) GetLimit() int {
	return int(of.pbFilter.GetLimit())
}
