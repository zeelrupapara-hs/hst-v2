package v1

import (
	"context"

	"hstserver/pkg/ws"
)

// CreateGroupWS creates a group from a socket command.
func (s *HttpServer) CreateGroupWS(c *ws.Ctx) error {
	var body CrtGroup
	if err := c.BodyParser(&body); err != nil {
		return c.SendError(err)
	}

	snap, err := s.SessionOf(c)
	if err != nil {
		return c.SendError(err)
	}

	v, _, err := s.MakeGroup(context.Background(), body, snap, c.Ip())
	if err != nil {
		return c.SendError(err)
	}

	return c.SendOK(v)
}

// UptGroupWS is an update sent over the socket, where the id travels in the payload.
type UptGroupWS struct {
	GroupID int `json:"group_id" validate:"required,gt=0"`
	UptGroup
}

// UpdateGroupWS patches a group from a socket command.
func (s *HttpServer) UpdateGroupWS(c *ws.Ctx) error {
	var body UptGroupWS
	if err := c.BodyParser(&body); err != nil {
		return c.SendError(err)
	}

	snap, err := s.SessionOf(c)
	if err != nil {
		return c.SendError(err)
	}

	v, _, err := s.ChangeGroup(context.Background(), body.GroupID, body.UptGroup, snap, c.Ip())
	if err != nil {
		return c.SendError(err)
	}

	return c.SendOK(v)
}

// DelGroupWS names the group to remove.
type DelGroupWS struct {
	GroupID int `json:"group_id" validate:"required,gt=0"`
}

// DeleteGroupWS removes a group from a socket command.
func (s *HttpServer) DeleteGroupWS(c *ws.Ctx) error {
	var body DelGroupWS
	if err := c.BodyParser(&body); err != nil {
		return c.SendError(err)
	}

	snap, err := s.SessionOf(c)
	if err != nil {
		return c.SendError(err)
	}

	if _, err := s.RemoveGroup(context.Background(), body.GroupID, snap, c.Ip()); err != nil {
		return c.SendError(err)
	}

	return c.SendOK(DelGroupWS{GroupID: body.GroupID})
}
