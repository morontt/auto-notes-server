package filters

import pb "xelbot.com/auto-notes/server/rpc/server"

type ServiceFilter struct {
	pbFilter *pb.ServiceFilter
	commonPart
}

func NewServiceFilter(f *pb.ServiceFilter) *ServiceFilter {
	return &ServiceFilter{
		pbFilter:   f,
		commonPart: commonPart{filter: f},
	}
}
