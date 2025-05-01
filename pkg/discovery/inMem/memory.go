package inMem

import (
	"context"
	"errors"
	"sync"
	"time"
)

type typeServiceName string
type typeInstanceId string

var serviceIsNotRegisterErr = errors.New("given service isn't found")
var instanceIsNotFoundErr = errors.New("given instance isn't found")

type Registry struct {
	sync.RWMutex
	serviceAddrs map[typeServiceName]map[typeInstanceId]*serviceInstance
}

func (r *Registry) Register(_ context.Context, instanceId, serviceName, hostPort string) error {
	r.RWMutex.Lock()
	defer r.RWMutex.Unlock()
	if _, ok := r.serviceAddrs[typeServiceName(serviceName)]; !ok {
		r.serviceAddrs[typeServiceName(serviceName)] = make(map[typeInstanceId]*serviceInstance)
	}
	r.serviceAddrs[typeServiceName(serviceName)][typeInstanceId(instanceId)] = &serviceInstance{
		hostPort: hostPort,
	}
	return nil
}

func (r *Registry) Deregister(ctx context.Context, instanceId, serviceName string) error {
	r.RWMutex.RUnlock()
	defer r.RWMutex.RUnlock()

	if _, ok := r.serviceAddrs[typeServiceName(serviceName)]; !ok {
		return serviceIsNotRegisterErr
	} else {
		if _, ok := r.serviceAddrs[typeServiceName(serviceName)][typeInstanceId(instanceId)]; !ok {
			return instanceIsNotFoundErr
		}
	}

	delete(r.serviceAddrs[typeServiceName(serviceName)], typeInstanceId(instanceId))
	return nil
}

func (r *Registry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	r.RWMutex.RLock()
	defer r.RWMutex.RUnlock()

	if len(r.serviceAddrs[typeServiceName(serviceName)]) == 0 {
		return nil, serviceIsNotRegisterErr
	}
	var res []string
	for _, i := range r.serviceAddrs[typeServiceName(serviceName)] {
		if i.lastActive.Before(time.Now().Add(-5 * time.Second)) {
			continue
		}
		res = append(res, i.hostPort)
	}

	return res, nil
}

func (r *Registry) ReportHealthyState(instanceId string, state string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[typeServiceName(instanceId)]; !ok {
		return serviceIsNotRegisterErr
	}

	if _, ok := r.serviceAddrs[typeServiceName(instanceId)][typeInstanceId(instanceId)]; !ok {
		return instanceIsNotFoundErr
	}

	r.serviceAddrs[typeServiceName(instanceId)][typeInstanceId(instanceId)].lastActive = time.Now()
	return nil
}

func NewRegistry() *Registry {
	return &Registry{
		serviceAddrs: make(map[typeServiceName]map[typeInstanceId]*serviceInstance),
	}
}

type serviceInstance struct {
	hostPort   string
	lastActive time.Time
}
