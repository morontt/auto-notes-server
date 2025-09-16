package filters

import pb "xelbot.com/auto-notes/server/rpc/server"

type FuelFilter struct {
	pbFilter *pb.FuelFilter
}

func NewFuelFilter(f *pb.FuelFilter) *FuelFilter {
	return &FuelFilter{pbFilter: f}
}

func (f *FuelFilter) GetPage() int {
	if f.pbFilter.GetPage() > 0 {
		return int(f.pbFilter.GetPage())
	}

	return 1
}

func (f *FuelFilter) GetLimit() int {
	return int(f.pbFilter.GetLimit())
}
