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

type pager struct {
	filter any
}

func (p *pager) GetPage() int {
	if pf, ok := p.filter.(pagerFilter); ok {
		return int(pf.GetPage())
	}

	return 1
}

func (p *pager) GetLimit() int {
	if pf, ok := p.filter.(pagerFilter); ok {
		return int(pf.GetLimit())
	}

	return 100
}
