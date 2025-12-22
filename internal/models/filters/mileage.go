package filters

import pb "xelbot.com/auto-notes/server/rpc/server"

type MileageFilter struct {
	pbFilter *pb.MileageFilter
}

func NewMileageFilter(f *pb.MileageFilter) *MileageFilter {
	return &MileageFilter{pbFilter: f}
}

func (mf *MileageFilter) GetPage() int {
	if mf.pbFilter.GetPage() > 0 {
		return int(mf.pbFilter.GetPage())
	}

	return 1
}

func (mf *MileageFilter) GetLimit() int {
	return int(mf.pbFilter.GetLimit())
}
