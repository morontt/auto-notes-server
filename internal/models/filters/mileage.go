package filters

import pb "xelbot.com/auto-notes/server/rpc/server"

type MileageFilter struct {
	pbFilter *pb.MileageFilter
	commonPart
}

func NewMileageFilter(f *pb.MileageFilter) *MileageFilter {
	return &MileageFilter{
		pbFilter:   f,
		commonPart: commonPart{filter: f},
	}
}
