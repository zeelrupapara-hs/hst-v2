package model

type QueryWhat int32

const (
	QueryAccount   QueryWhat = 1
	QueryPositions QueryWhat = 2
	QueryOrders    QueryWhat = 3
	QueryState     QueryWhat = 4
)

type QueryRequest struct {
	Login int64     `json:"login"`
	What  QueryWhat `json:"what"`
}

type QueryResult struct {
	Login     int64      `json:"login"`
	Found     bool       `json:"found"`
	Account   *Account   `json:"account,omitempty"`
	Positions []Position `json:"positions,omitempty"`
	Orders    []Order    `json:"orders,omitempty"`
}
