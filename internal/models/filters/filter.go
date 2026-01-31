package filters

type PaginationPart interface {
	GetPage() int
	GetLimit() int
}

func GetLastPage(f PaginationPart, cntItems int) int {
	var lastPage = 1
	if f.GetLimit() > 0 {
		lastPage += int((float32(cntItems) - 0.5) / float32(f.GetLimit()))
	}

	return lastPage
}

type pagerFilter interface {
	GetPage() int32
	GetLimit() int32
}

type carPart interface {
	GetCarId() int32
}

type commonPart struct {
	filter any
}

func (p *commonPart) GetPage() int {
	if pf, ok := p.filter.(pagerFilter); ok {
		if pf.GetPage() > 0 {
			return int(pf.GetPage())
		}
	}

	return 1
}

func (p *commonPart) GetLimit() int {
	if pf, ok := p.filter.(pagerFilter); ok {
		return int(pf.GetLimit())
	}

	return 100
}

func (p *commonPart) HasCarId() bool {
	if pf, ok := p.filter.(carPart); ok {
		return pf.GetCarId() > 0
	}

	return false
}

func (p *commonPart) GetCarId() uint {
	if pf, ok := p.filter.(carPart); ok {
		return uint(pf.GetCarId())
	}

	return 0
}
