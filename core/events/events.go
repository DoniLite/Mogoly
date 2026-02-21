package events

import (
	goevents "github.com/DoniLite/go-events"
)

var eventBus *goevents.EventFactory

var (
	ServerStartedEvent     *goevents.Event
	ErrorDroppedEvent      *goevents.Event
	CertManagerActionEvent *goevents.Event
	ConfigFileUpdateEvent  *goevents.Event
)

func init() {
	eventBus = goevents.NewEventBus()

	ServerStartedEvent = eventBus.CreateEvent("server_started")
	ErrorDroppedEvent = eventBus.CreateEvent("error_dropped")
	CertManagerActionEvent = eventBus.CreateEvent("cert_manager_action")
	ConfigFileUpdateEvent = eventBus.CreateEvent("config_file_update")
}

func AddEventHandler(event *goevents.Event, handler goevents.EventHandler) {
	Logf(LOG_INFO, "[EVENT]: New registered event handler for the %s event", event.Name)
	eventBus.On(event, handler)
}

func SubscribeToEvent(handler goevents.EventHandler, events ...*goevents.Event) {
	Logf(LOG_INFO, "[EVENT]: New subscriber to %v events", events)
	eventBus.Subscribe(handler, events...)
}

func GetEventBus() *goevents.EventFactory {
	return eventBus
}
