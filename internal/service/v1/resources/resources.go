package resources

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/kube-orchestra/maestro/db"
	maestroMqtt "github.com/kube-orchestra/maestro/internal/mqtt"
	v1 "github.com/kube-orchestra/maestro/proto/api/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

type ResourcesService struct {
	v1.UnimplementedResourceServiceServer
	resourceChan chan<- maestroMqtt.ResourceMessage
}

func NewResourceService(resourceChan chan<- maestroMqtt.ResourceMessage) *ResourcesService {
	return &ResourcesService{resourceChan: resourceChan}
}

func (svc *ResourcesService) Read(_ context.Context, r *v1.ResourceReadRequest) (*v1.Resource, error) {
	res, err := db.GetResource(r.Id)
	if err != nil {
		return nil, err
	}

	objStructpb, err := structpb.NewStruct(res.Object.UnstructuredContent())
	if err != nil {
		return nil, err
	}

	resResponse := &v1.Resource{
		Id:         res.Id,
		ConsumerId: res.ConsumerId,
		Object:     objStructpb,
	}
	return resResponse, nil
}

func (svc *ResourcesService) Create(_ context.Context, r *v1.ResourceCreateRequest) (*v1.Resource, error) {
	unstructuredObject := unstructured.Unstructured{Object: r.Object.AsMap()}
	res := db.Resource{
		Id:         uuid.NewString(),
		ConsumerId: r.ConsumerId,
		Object:     unstructuredObject,
	}

	// TODO: check that it doesn't exist
	err := db.PutResource(&res)
	if err != nil {
		return nil, err
	}

	messageMeta := maestroMqtt.MessageMeta{
		Id:                   res.Id,
		ConsumerId:           res.ConsumerId,
		SentTimestamp:        0,
		ResourceGenerationID: "resId",
	}
	resourceMessage := maestroMqtt.ResourceMessage{
		MessageMeta: messageMeta,
		Content:     &unstructuredObject,
	}
	svc.resourceChan <- resourceMessage

	return &v1.Resource{Id: res.Id, ConsumerId: res.ConsumerId, Object: r.Object}, nil
}

func (svc *ResourcesService) Update(_ context.Context, r *v1.ResourceUpdateRequest) (*v1.Resource, error) {
	// check that it exists
	res, err := db.GetResource(r.Id)
	if err != nil {
		return nil, err
	}

	res.Object = unstructured.Unstructured{Object: r.Object.AsMap()}

	// TODO: increment generation

	messageMeta := maestroMqtt.MessageMeta{
		Id:                   res.Id,
		ConsumerId:           res.ConsumerId,
		SentTimestamp:        0,
		ResourceGenerationID: "resId",
	}
	resourceMessage := maestroMqtt.ResourceMessage{
		MessageMeta: messageMeta,
		Content:     &res.Object,
	}
	svc.resourceChan <- resourceMessage

	return &v1.Resource{Id: res.Id, ConsumerId: res.ConsumerId, Object: r.Object}, nil
}
