package addons

import (
	"context"
	v1 "github.com/kube-orchestra/maestro/proto/api/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ConsumersRegistrationService struct {
	v1.UnimplementedConsumerRegistrationServer
}

func NewConsumerRegistrationService() *ConsumersRegistrationService {
	return &ConsumersRegistrationService{}
}

func (svc *ConsumersRegistrationService) List(_ context.Context, _ *emptypb.Empty) (*v1.ConsumerList, error) {
	c := &v1.ConsumerList{
		Consumers: []*v1.Consumer{
			{
				Id:     "1",
				Name:   "Foo",
				Labels: nil,
			},
			{
				Id:     "2",
				Name:   "Bar",
				Labels: nil,
			},
		},
	}
	return c, nil
}

func (svc *ConsumersRegistrationService) Read(_ context.Context, r *v1.ConsumerReadRequest) (*v1.Consumer, error) {
	c := &v1.Consumer{
		Id:     r.Id,
		Name:   "Foo",
		Labels: nil,
	}
	return c, nil
}
