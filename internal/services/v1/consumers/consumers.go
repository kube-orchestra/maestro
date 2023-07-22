package consumers

import (
	"context"
	"github.com/google/uuid"
	"github.com/kube-orchestra/maestro/db"
	v1 "github.com/kube-orchestra/maestro/proto/api/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ConsumersService struct {
	v1.UnimplementedConsumerServiceServer
}

func NewConsumerService() *ConsumersService {
	return &ConsumersService{}
}

// TODO remove
func (svc *ConsumersService) List(_ context.Context, _ *emptypb.Empty) (*v1.ConsumerList, error) {
	c, err := db.ListConsumers()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (svc *ConsumersService) Read(_ context.Context, r *v1.ConsumerReadRequest) (*v1.Consumer, error) {
	c, err := db.GetConsumer(r.Id)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (svc *ConsumersService) Create(_ context.Context, r *v1.Consumer) (*v1.Consumer, error) {
	r.Id = uuid.NewString()

	consumerStruct := db.ConvertProtobufToStruct(r)

	err := db.PutConsumer(consumerStruct)
	if err != nil {
		return nil, err
	}
	return r, nil
}
