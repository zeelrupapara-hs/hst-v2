package handler

// Every subject this service publishes or consumes belongs here, as a
// constant. A subject typed inline in two files is a subject that will be
// renamed in one of them.
//
// Convention: <service>.<entity>.<event>, with a token per identifier so a
// consumer can follow one entity instead of filtering the whole stream.
//
//	hstcore.example.created
//	hstcore.example.updated.*
//
// Queue groups go beside them: every instance subscribes with the same group
// name, so exactly one of them handles each message.
const (
	// SubjectExample is a placeholder, replace it with the real ones.
	SubjectExample = "hstcore.example.*"

	// GroupExample is the queue group for SubjectExample.
	GroupExample = "hstcore-example"
)
