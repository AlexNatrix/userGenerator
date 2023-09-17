package sort

import (
	"context"
	"net/http"
	"strings"
)

const (
	ascSort = "ASC"
	descSort = "DESC"
	defaultSortPararm = "ID"
)

type OptionsContextKey string;

type Options struct{
	Field, Order string
}

func Middleware(h http.HandlerFunc) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		sortBy := r.URL.Query().Get("sort_by")
		sortOrder := r.URL.Query().Get("sort_order")
		if sortBy==""{
			sortBy=defaultSortPararm
		}
		upperCaseSort:=strings.ToUpper(sortOrder)
		if upperCaseSort!=ascSort && upperCaseSort!=descSort {
			sortOrder=ascSort
		}
		opt := Options{
			Field: sortBy,
			Order: sortOrder,
		}
		ctx:=context.WithValue(r.Context(),OptionsContextKey("sort_options"),opt)
		r = r.WithContext(ctx)
		h(w,r)	
	}
}