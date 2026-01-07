package filters

import pb "xelbot.com/auto-notes/server/rpc/server"

type ServiceFilter struct {
	pbFilter *pb.ServiceFilter
	pager
}

func NewServiceFilter(f *pb.ServiceFilter) *ServiceFilter {
	return &ServiceFilter{
		pbFilter: f,
		pager:    pager{filter: f},
	}
}
