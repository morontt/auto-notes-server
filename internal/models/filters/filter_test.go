package filters

import (
	"strconv"
	"testing"
)

type filterDumb struct{}

func (f *filterDumb) GetPage() int {
	return 1
}

func (f *filterDumb) GetLimit() int {
	return 3
}

func TestBotFilter(t *testing.T) {
	tests := []struct {
		page, want int
	}{
		{
			page: 1,
			want: 1,
		},
		{
			page: 2,
			want: 1,
		},
		{
			page: 3,
			want: 1,
		},
		{
			page: 4,
			want: 2,
		},
		{
			page: 5,
			want: 2,
		},
		{
			page: 6,
			want: 2,
		},
		{
			page: 7,
			want: 3,
		},
	}

	for idx, item := range tests {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			f := &filterDumb{}
			res := GetLastPage(f, item.page)

			if res != item.want {
				t.Errorf("%d : got %d; want %d", item.page, res, item.want)
			}
		})
	}
}
