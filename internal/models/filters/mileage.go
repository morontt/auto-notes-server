package filters

import pb "xelbot.com/auto-notes/server/rpc/server"

type MileageFilter struct {
	pbFilter *pb.MileageFilter
	pager
}

func NewMileageFilter(f *pb.MileageFilter) *MileageFilter {
	return &MileageFilter{
		pbFilter: f,
		pager:    pager{filter: f},
	}
}
