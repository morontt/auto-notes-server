package filters

import pb "xelbot.com/auto-notes/server/rpc/server"

type FuelFilter struct {
	pbFilter *pb.FuelFilter
	commonPart
}

func NewFuelFilter(f *pb.FuelFilter) *FuelFilter {
	return &FuelFilter{
		pbFilter:   f,
		commonPart: commonPart{filter: f},
	}
}
