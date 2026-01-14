package router

import (
	"github.com/DoniLite/Mogoly/cloud"
	"github.com/DoniLite/Mogoly/core/events"
)

var manager *cloud.CloudManager

func init() {
	var err error
	manager, err = cloud.NewCloudDBManager()
	if err != nil {
		panic(err)
	}
}

func GetServices() []*cloud.ServiceInstance {
	return manager.ListInstances()
}

func AddService(service cloud.ServiceConfig) {
	_, err := manager.CreateInstance(service)
	if err != nil {
		events.Logf(events.LOG_ERROR, "[SERVICE]: Error while adding service: %v", err)
		return
	}
	events.Logf(events.LOG_INFO, "[SERVICE]: Service %s added successfully", service.Name)
}