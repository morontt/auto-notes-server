package filters

import pb "xelbot.com/auto-notes/server/rpc/server"

type FuelFilter struct {
	pbFilter *pb.FuelFilter
	pager
}

func NewFuelFilter(f *pb.FuelFilter) *FuelFilter {
	return &FuelFilter{
		pbFilter: f,
		pager:    pager{filter: f},
	}
}
