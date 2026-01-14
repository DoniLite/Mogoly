package events

import (
	"context"

	goevents "github.com/DoniLite/go-events"
)

func OnCertManagerEvent(ctx context.Context, event string, data map[string]any) error {
	Logf(LOG_INFO, "[EVENT]: Event emitting certmanager event %s with data: %v", event, data)
	eventBus.Emit(CertManagerActionEvent, &goevents.EventData{Message: event, Payload: data})
	return nil
}
