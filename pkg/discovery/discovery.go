package discovery

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type Registry interface {
	Register(ctx context.Context, instanceId, serviceName, hostPort string) error
	Deregister(ctx context.Context, instanceId, serviceName string) error
	ServiceAddresses(ctx context.Context, serviceName string) ([]string, error)
	ReportHealthyState(instanceId string, state string) error
}

func GenerateInstanceID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}
