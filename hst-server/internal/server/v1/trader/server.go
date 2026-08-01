// Package trader is the trading account panel: what an account may see about itself.
package trader

import (
	v1 "hstserver/internal/server/v1"
)

// Server is the shared core, acting for a trading account.
type Server struct {
	*v1.HttpServer
}

// New wraps the core so the trader handlers can hang off it.
func New(core *v1.HttpServer) *Server { return &Server{core} }

// Response is the envelope every endpoint returns, aliased here for swagger.
type Response = v1.Response

// ViewSymbol is the shared instrument summary, aliased for the same reason.
type ViewSymbol = v1.ViewSymbol
