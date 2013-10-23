package models

import (
	"fmt"
	"html/template"
	"net/url"
)

type SimplePaginator struct {
	HasPrev     bool
	HasNext     bool
	CurrentPage int
	TotalPages  int
	OtherParams template.URL
	PrevPage    int
	NextPage    int
}

func NewSimplePaginator(currentPage int, total int, numInOnePage int, params url.Values) *SimplePaginator {
	paginator := &SimplePaginator{}
	paginator.CurrentPage = currentPage
	paginator.TotalPages = total / numInOnePage
	params.Del("p")
	paginator.OtherParams = template.URL(params.Encode())
	fmt.Println(paginator.OtherParams)
	if total%numInOnePage > 0 {
		paginator.TotalPages += 1
	}
	if paginator.CurrentPage > 1 {
		paginator.HasPrev = true
		paginator.PrevPage = paginator.CurrentPage - 1
	}
	if paginator.CurrentPage < paginator.TotalPages {
		paginator.HasNext = true
		paginator.NextPage = paginator.CurrentPage + 1
	}
	return paginator
}
