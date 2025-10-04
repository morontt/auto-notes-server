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
