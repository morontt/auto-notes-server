package filters

import pb "xelbot.com/auto-notes/server/rpc/server"

type OrderFilter struct {
	pbFilter *pb.OrderFilter
	commonPart
}

func NewOrderFilter(f *pb.OrderFilter) *OrderFilter {
	return &OrderFilter{
		pbFilter:   f,
		commonPart: commonPart{filter: f},
	}
}

func (p *OrderFilter) GetTypeId() int32 {
	if p != nil {
		return p.pbFilter.GetTypeId()
	}

	return 0
}
