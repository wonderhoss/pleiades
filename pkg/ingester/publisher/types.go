package publisher

import "github.com/gargath/pleiades/pkg/ingester/sse"

//Publisher defines ways to send received Events to other systems or process them in some form
type Publisher interface {
	ReadAndPublish() (int64, error)
	ProcessEvent(*sse.Event) error
	GetResumeID() string
	ValidateConnection() error
}
