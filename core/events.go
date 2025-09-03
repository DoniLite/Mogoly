package core

import (
	"context"

	goevents "github.com/DoniLite/go-events"
)

var EventBus *goevents.EventFactory

var (
	ServerStartedEvent     *goevents.Event
	ErrorDroppedEvent      *goevents.Event
	CertManagerActionEvent *goevents.Event
)

func init() {
	EventBus = goevents.NewEventBus()

	ServerStartedEvent = EventBus.CreateEvent("server_started")
	ErrorDroppedEvent = EventBus.CreateEvent("error_dropped")
	CertManagerActionEvent = EventBus.CreateEvent("cert_manager_action")
}

func AddEventHandler(event *goevents.Event, handler goevents.EventHandler) {
	logf(LOG_INFO, "[EVENT]: New registered event handler for the %s event", event.Name)
	EventBus.On(event, handler)
}

func onCertManagerEvent(ctx context.Context, event string, data map[string]any) error {
	logf(LOG_INFO, "[EVENT]: Event emitting certmanager event %s with data: %v", event, data)
	EventBus.Emit(CertManagerActionEvent, &goevents.EventData{Message: event, Payload: data})
	return nil
}

func SubscribeToEvent(handler goevents.EventHandler, events ...*goevents.Event) {
	logf(LOG_INFO, "[EVENT]: New subscriber to %v events", events)
	EventBus.Subscribe(handler, events...)
}
