package dtoadapter

import (
	"github.com/xsqrty/notes/internal/domain/search"
	"github.com/xsqrty/op/orm"
)

func SearchToPaginateRequest(req *search.Request) *orm.PaginateRequest {
	return &orm.PaginateRequest{
		Orders:  searchOrdersToPaginate(req.Orders),
		Filters: orm.PaginateFilters(req.Filters),
		Limit:   req.Limit,
		Offset:  req.Offset,
	}
}

func searchOrdersToPaginate(orders []search.Order) []orm.PaginateOrder {
	res := make([]orm.PaginateOrder, len(orders))
	for i := range orders {
		res[i] = orm.PaginateOrder{Key: orders[i].Key, Desc: orders[i].Desc}
	}

	return res
}
