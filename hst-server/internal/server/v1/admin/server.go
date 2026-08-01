// Package admin is the manager panel: everything a member of staff may configure.
package admin

import (
	v1 "hstserver/internal/server/v1"
)

// Server is the shared core, acting for a manager.
type Server struct {
	*v1.HttpServer
}

// New wraps the core so the manager handlers can hang off it.
func New(core *v1.HttpServer) *Server { return &Server{core} }

// Response is the envelope every endpoint returns, aliased here for swagger.
type Response = v1.Response
