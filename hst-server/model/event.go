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
	// FormatBinary forwards the bytes verbatim in a binary frame. Nothing is
	// parsed, copied or re-encoded, which is what a tick stream needs.
	FormatBinary
	// FormatText forwards the bytes verbatim in a text frame.
	FormatText
)

// Event is what crosses nats and what a websocket client receives.
//
// The payload stays raw the whole way through. This server routes events, it
// does not need to understand the bodies, and decoding every message to
// re-encode it would cost more than the delivery itself.
type Event struct {
	Type string `json:"type"`
	// Payload is whatever the publisher put in, verbatim. It is only valid
	// json when Format is FormatJSON.
	Payload []byte `json:"-"`
	// At is unix nanoseconds.
	At int64 `json:"at,omitempty"`
	// Format decides the frame the client receives. It never crosses the wire
	// itself: a binary frame is its own answer.
	Format Format `json:"-"`
}

// Header keys a publisher can set on the nats message to say how the payload
// should be delivered. Without them the payload is sniffed, so an existing
// publisher that sends json needs no changes.
const (
	// HeaderFormat is "json", "binary" or "text".
	HeaderFormat = "X-Format"
	// HeaderContentType is honoured as well, since a publisher may already be
	// setting it: application/json, text/plain, anything else is binary.
	HeaderContentType = "Content-Type"
)

// FormatFromHeader reads the declared format. ok is false when the publisher
// declared nothing and the payload has to be sniffed instead.
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
// Only json is worth detecting: it is the one format that has to be wrapped,
// and anything else is safer sent through untouched.
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

// Event types. One constant per event, so a subject typed in two files cannot
// drift apart.
const (
	EventPing    = "ping"
	EventPong    = "pong"
	EventError   = "error"
	EventWelcome = "welcome"

	// EventSessionRevoked tells a socket its session is gone; the server
	// closes the connection right after sending it.
	EventSessionRevoked = "session.revoked"
)

// Subject roots. The websocket layer never subscribes to anything outside
// these, so a publisher cannot reach a socket by accident.
const (
	// SubjectSessionRoot targets exactly one session.
	SubjectSessionRoot = "ws.session"
	// SubjectLoginRoot targets every session of one login, which is what a
	// user with two terminals open has.
	SubjectLoginRoot = "ws.login"
	// SubjectRightRoot targets every manager holding one right. This is the
	// authorisation boundary: a socket only ever subscribes to the rights its
	// session actually has, so an unauthorised event never reaches it.
	SubjectRightRoot = "ws.right"
	// SubjectBroadcastRoot reaches every connected socket.
	SubjectBroadcastRoot = "ws.broadcast"
)

// SubjectSession is the subject one session listens on: ws.session.<sid>.>
//
// The wildcard is ">" and not "*" on purpose: "*" matches exactly one token,
// so an event named "symbol.updated" would never be delivered. Event names are
// hierarchical, and the grammar has to allow that.
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

// EventTypeFromSubject recovers the event name from the subject, so a
// publisher does not have to repeat it inside the payload.
//
// Everything after the routing prefix is the name, dots included:
// ws.right.journals.entry.created is the event "entry.created".
func EventTypeFromSubject(subject string) string {
	parts := strings.Split(subject, ".")

	// ws.broadcast.<event...> carries no identifier, the others do
	skip := 3
	if len(parts) >= 2 && parts[0]+"."+parts[1] == SubjectBroadcastRoot {
		skip = 2
	}

	if len(parts) <= skip {
		return ""
	}
	return strings.Join(parts[skip:], ".")
}
