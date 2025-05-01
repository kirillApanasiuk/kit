package counsul

import (
	"context"
	"errors"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"github.com/kirillApanasiuk/kit/pkg/discovery"
	"strconv"
	"strings"
)

type Registry struct {
	client *consul.Client
}

func (r *Registry) Register(ctx context.Context, instanceId, serviceName, hostPort string) error {
	parts := strings.Split(hostPort, ":")
	if len(parts) != 2 {
		return errors.New("hostPort must be in a form of <host>:<port>, example: localhost:8001")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}
	return r.client.Agent().ServiceRegister(&consul.AgentServiceRegistration{
		Address: parts[0],
		ID:      instanceId,
		Name:    serviceName,
		Port:    port,
		Check: &consul.AgentServiceCheck{
			CheckID: instanceId,
			TTL:     "5s",
		},
	})
}

func (r *Registry) Deregister(ctx context.Context, instanceId, serviceName string) error {
	return r.client.Agent().ServiceDeregister(instanceId)
}

func (r *Registry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	entries, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	} else if len(entries) == 0 {
		return nil, discovery.ErrNotFound
	}

	res := make([]string, 0, len(entries))
	for _, entry := range entries {
		hostPort := fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port)
		res = append(res, hostPort)

	}

	return res, nil
}

func (r *Registry) ReportHealthyState(instanceId string, state string) error {
	return r.client.Agent().PassTTL(instanceId, "")
}

func NewRegistry(addr string) (*Registry, error) {
	config := consul.DefaultConfig()
	config.Address = addr
	client, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Registry{client}, nil
}
