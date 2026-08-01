package model

import (
	"fmt"
	"strings"
)

// Format is how a payload is carried to the client.
type Format uint8

const (
	// FormatJSON wraps the payload in the event envelope. The default.
	FormatJSON Format = iota
	// FormatBinary forwards the bytes verbatim in a binary frame.
	FormatBinary
	// FormatText forwards the bytes verbatim in a text frame.
	FormatText
)

// Event is what crosses nats and what a websocket client receives.
type Event struct {
	Type string `json:"type"`
	// Payload is whatever the publisher put in, verbatim.
	Payload []byte `json:"-"`
	// Group is the group path the record belongs to, for a group scoped event.
	Group string `json:"group,omitempty"`
	// At is unix nanoseconds.
	At int64 `json:"at,omitempty"`
	// Format decides the frame the client receives.
	Format Format `json:"-"`
}

// Header keys a publisher can set on the nats message to say how the payload should be delivered.
const (
	// HeaderFormat is "json", "binary" or "text".
	HeaderFormat = "X-Format"
	// Payload is whatever the publisher put in, valid json only when Format is FormatJSON
	HeaderContentType = "Content-Type"
	// HeaderEvent carries what happened to the record: created, updated, deleted, moved.
	HeaderEvent = "X-Event"
)

// FormatFromHeader reads the declared format.
func FormatFromHeader(format, contentType string) (Format, bool) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return FormatJSON, true
	case "binary", "bin":
		return FormatBinary, true
	case "text", "txt":
		return FormatText, true
	}

	ct := strings.ToLower(strings.TrimSpace(contentType))
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	switch ct {
	case "":
		return FormatJSON, false
	case "application/json", "text/json":
		return FormatJSON, true
	case "text/plain":
		return FormatText, true
	default:
		return FormatBinary, true
	}
}

// SniffFormat guesses from the payload when the publisher declared nothing.
func SniffFormat(payload []byte) Format {
	for _, b := range payload {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '{', '[':
			return FormatJSON
		default:
			return FormatBinary
		}
	}
	return FormatBinary
}

// Event types.
const (
	EventPing    = "ping"
	EventPong    = "pong"
	EventError   = "error"
	EventWelcome = "welcome"

	// EventSessionRevoked tells a socket its session is gone; the server closes the connection right after sending it.
	EventSessionRevoked = "session.revoked"

	EventOrderCreate       = "order_create"
	EventOrderDealerCreate = "order_dealer_create"
	EventOrderUpdate       = "order_update"
	EventOrderDealerUpdate = "order_dealer_update"
	EventOrderCancel       = "order_cancel"
	EventOrderDealerCancel = "order_dealer_cancel"

	EventPositionUpdate       = "position_update"
	EventPositionDealerUpdate = "position_dealer_update"
	EventPositionClose        = "position_close"
	EventPositionDealerClose  = "position_dealer_close"
	EventPositionCloseBy      = "position_close_by"

	EventDealerConfirm = "dealer_confirm"
	EventDealerRequote = "dealer_requote"
	EventDealerReject  = "dealer_reject"
	EventDealerCancel  = "dealer_cancel"

	EventBadRequest          = "bad_request"
	EventNotFound            = "not_found"
	EventForbidden           = "forbidden"
	EventUnauthorized        = "unauthorized"
	EventInternalServerError = "internal_server_error"
)

// ErrorPayload is what a refused socket request carries back.
type ErrorPayload struct {
	Message string `json:"message"`
	Reason  string `json:"reason"`
}

// Subject roots.
const (
	// SubjectSessionRoot targets exactly one session.
	SubjectSessionRoot = "ws.session"
	// SubjectLoginRoot targets every session of one login, which is what a user with two terminals open has.
	SubjectLoginRoot = "ws.login"
	// SubjectRightRoot targets every manager holding one right.
	SubjectRightRoot = "ws.right"
	// SubjectBroadcastRoot reaches every connected socket.
	SubjectBroadcastRoot = "ws.broadcast"
)

// everything after the routing prefix is the name, dots included
func SubjectSession(sid string) string {
	return fmt.Sprintf("%s.%s.>", SubjectSessionRoot, sid)
}

// SubjectSessionEvent addresses one event at one session.
func SubjectSessionEvent(sid, event string) string {
	return fmt.Sprintf("%s.%s.%s", SubjectSessionRoot, sid, event)
}

// SubjectLogin is every session of one login: ws.login.<login>.>
func SubjectLogin(login int64) string {
	return fmt.Sprintf("%s.%d.>", SubjectLoginRoot, login)
}

// SubjectLoginEvent addresses one event at every session of one login.
func SubjectLoginEvent(login int64, event string) string {
	return fmt.Sprintf("%s.%d.%s", SubjectLoginRoot, login, event)
}

// SubjectRight is every manager holding one right: ws.right.<right>.>
func SubjectRight(right string) string {
	return fmt.Sprintf("%s.%s.>", SubjectRightRoot, right)
}

// SubjectRightEvent addresses one event at every manager holding one right.
func SubjectRightEvent(right, event string) string {
	return fmt.Sprintf("%s.%s.%s", SubjectRightRoot, right, event)
}

// SubjectBroadcast is every connected socket.
func SubjectBroadcast() string {
	return SubjectBroadcastRoot + ".>"
}

// SubjectBroadcastEvent addresses one event at every connected socket.
func SubjectBroadcastEvent(event string) string {
	return SubjectBroadcastRoot + "." + event
}

// EventTypeFromSubject recovers the event name from the subject, so a publisher does not have to repeat it inside the payload.
func EventTypeFromSubject(subject string) string {
	typ, _ := ParseSubject(subject)
	return typ
}

// ParseSubject splits a subject into the event type and, for a group scoped subject, the group path it happened in.
func ParseSubject(subject string) (eventType, groupPath string) {
	parts := strings.Split(subject, ".")

	// ws.g.<family>.<path...>: everything after the family is the group path, and the event type is not in the subject at all.
	if len(parts) >= 4 && parts[0]+"."+parts[1] == SubjectGroupRoot {
		return parts[2], strings.Join(parts[3:], GroupSep)
	}

	// ws.broadcast.<event...> carries no identifier, the others do
	skip := 3
	if len(parts) >= 2 && parts[0]+"."+parts[1] == SubjectBroadcastRoot {
		skip = 2
	}
	if len(parts) <= skip {
		return "", ""
	}
	return strings.Join(parts[skip:], "."), ""
}
