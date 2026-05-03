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

func (p *FuelFilter) GetTypeId() int32 {
	if p != nil {
		return p.pbFilter.GetTypeId()
	}

	return 0
}

func (p *FuelFilter) GetStationId() int32 {
	if p != nil {
		return p.pbFilter.GetStationId()
	}

	return 0
}
