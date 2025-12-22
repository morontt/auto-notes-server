package filters

import pb "xelbot.com/auto-notes/server/rpc/server"

type OrderFilter struct {
	pbFilter *pb.OrderFilter
	pager
}

func NewOrderFilter(f *pb.OrderFilter) *OrderFilter {
	return &OrderFilter{
		pbFilter: f,
		pager:    pager{filter: f},
	}
}
