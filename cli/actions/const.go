package actions

import "github.com/DoniLite/Mogoly/sync"

// Action types for daemon commands
// Starting from 100 to avoid conflict with sync package internal actions
const (
	// Cloud actions

	ActionCloudCreate sync.Action_Type = 100 + iota
	ActionCloudList
	ActionCloudStart
	ActionCloudStop
	ActionCloudRestart
	ActionCloudDelete
	ActionCloudLogs
	ActionCloudInspect

	// Load balancer actions

	ActionLBCreate
	ActionLBList
	ActionLBAddBackend
	ActionLBRemoveBackend
	ActionLBHealth
	ActionLBStart
	ActionLBStop

	// Daemon actions

	ActionDaemonStatus
	ActionDaemonLogs
	ActionDaemonPing

	// Domain actions

	ActionDomainAdd
	ActionDomainList
	ActionDomainRemove
)
