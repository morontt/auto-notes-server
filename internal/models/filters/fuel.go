package filters

import pb "xelbot.com/auto-notes/server/rpc/server"

type FuelFilter struct {
	pbFilter *pb.FuelFilter
}

func NewFuelFilter(f *pb.FuelFilter) *FuelFilter {
	return &FuelFilter{pbFilter: f}
}

func (ff *FuelFilter) GetPage() int {
	if ff.pbFilter.GetPage() > 0 {
		return int(ff.pbFilter.GetPage())
	}

	return 1
}

func (ff *FuelFilter) GetLimit() int {
	return int(ff.pbFilter.GetLimit())
}
