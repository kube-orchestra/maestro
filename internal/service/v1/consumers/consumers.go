package consumers

import (
	"context"

	"github.com/google/uuid"
	"github.com/kube-orchestra/maestro/internal/db"
	v1 "github.com/kube-orchestra/maestro/proto/api/v1"
)

type Service struct {
	v1.UnimplementedConsumerServiceServer
}

func NewConsumerService() *Service {
	return &Service{}
}

func (svc *Service) Read(_ context.Context, r *v1.ConsumerReadRequest) (*v1.Consumer, error) {
	c, err := db.GetConsumer(r.Id)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type ConsumerExistsError struct{}

func (m *ConsumerExistsError) Error() string {
	return "Consumer already exists, use method PUT to update"
}

func (svc *Service) Create(_ context.Context, r *v1.ConsumerCreateRequest) (*v1.Consumer, error) {
	if r.Id != "" {
		c, err := db.GetConsumer(r.Id)
		if err != nil {
			return nil, err
		}
		if c != nil {
			return nil, &ConsumerExistsError{}
		}
	}

	newConsumer := &v1.Consumer{
		Id:     uuid.NewString(),
		Labels: r.Labels,
	}

	err := db.PutConsumer(newConsumer)
	if err != nil {
		return nil, err
	}

	return newConsumer, nil
}

type ConsumerDoesNotExistError struct{}

func (m *ConsumerDoesNotExistError) Error() string {
	return "Consumer doesn't exist, use method create it with method POST first"
}

func (svc *Service) Update(_ context.Context, c *v1.ConsumerUpdateRequest) (*v1.Consumer, error) {
	consumer, err := db.GetConsumer(c.Id)
	if err != nil {
		return nil, err
	}
	if consumer == nil {
		return nil, &ConsumerDoesNotExistError{}
	}

	updatedConsumer := &v1.Consumer{
		Id:     c.Id,
		Labels: c.Labels,
	}

	err = db.PutConsumer(updatedConsumer)
	if err != nil {
		return nil, err
	}

	return updatedConsumer, nil
}
