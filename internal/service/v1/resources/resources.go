package resources

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/kube-orchestra/maestro/internal/db"
	v1 "github.com/kube-orchestra/maestro/proto/api/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

type ResourcesService struct {
	v1.UnimplementedResourceServiceServer
	resourceChan chan<- db.ResourceMessage
}

func NewResourceService(resourceChan chan<- db.ResourceMessage) *ResourcesService {
	return &ResourcesService{resourceChan: resourceChan}
}

func (svc *ResourcesService) Read(_ context.Context, r *v1.ResourceReadRequest) (*v1.Resource, error) {
	res, err := db.GetResource(r.Id)
	if err != nil {
		return nil, err
	}

	// object to proto struct
	objProtoStruct, err := structpb.NewStruct(res.Object.UnstructuredContent())
	if err != nil {
		return nil, err
	}

	// status to proto struct
	statusJson, _ := json.Marshal(&res.Status)
	var statusMap map[string]interface{}
	err = json.Unmarshal(statusJson, &statusMap)
	if err != nil {
		return nil, err
	}
	statusProtoStruct, err := structpb.NewStruct(statusMap)
	if err != nil {
		return nil, err
	}

	resResponse := &v1.Resource{
		Id:           res.Id,
		ConsumerId:   res.ConsumerId,
		GenerationId: res.ResourceGenerationID,
		Object:       objProtoStruct,
		Status:       statusProtoStruct,
	}

	return resResponse, nil
}

func (svc *ResourcesService) Create(_ context.Context, r *v1.ResourceCreateRequest) (*v1.Resource, error) {
	unstructuredObject := unstructured.Unstructured{Object: r.Object.AsMap()}

	// set uid
	uid := uuid.NewString()
	unstructuredObject.SetUID(types.UID(uid))

	res := db.Resource{
		Id:                   uid,
		ConsumerId:           r.ConsumerId,
		Object:               unstructuredObject,
		ResourceGenerationID: 1,
	}

	// TODO: check that it doesn't exist
	err := db.PutResource(&res)
	if err != nil {
		return nil, err
	}

	messageMeta := db.MessageMeta{
		SentTimestamp:        0,
		ResourceGenerationID: res.ResourceGenerationID,
	}
	resourceMessage := db.ResourceMessage{
		Id:          res.Id,
		ConsumerId:  res.ConsumerId,
		MessageMeta: messageMeta,
		Content:     &unstructuredObject,
	}
	svc.resourceChan <- resourceMessage

	return &v1.Resource{Id: res.Id,
		ConsumerId:   res.ConsumerId,
		GenerationId: res.ResourceGenerationID,
		Object:       r.Object}, nil
}

func (svc *ResourcesService) Update(_ context.Context, r *v1.ResourceUpdateRequest) (*v1.Resource, error) {
	// TODO: rewrite using UpdateItem dynamodb

	// check that it exists
	res, err := db.GetResource(r.Id)
	if err != nil {
		return nil, err
	}

	res.Object = unstructured.Unstructured{Object: r.Object.AsMap()}
	res.Object.SetUID(types.UID(r.Id))
	res.ResourceGenerationID++

	err = db.PutResource(res)
	if err != nil {
		return nil, err
	}

	messageMeta := db.MessageMeta{
		SentTimestamp:        0,
		ResourceGenerationID: res.ResourceGenerationID + 1,
	}

	resourceMessage := db.ResourceMessage{
		Id:          res.Id,
		ConsumerId:  res.ConsumerId,
		MessageMeta: messageMeta,
		Content:     &res.Object,
	}
	svc.resourceChan <- resourceMessage

	return &v1.Resource{Id: res.Id,
		ConsumerId:   res.ConsumerId,
		GenerationId: res.ResourceGenerationID,
		Object:       r.Object}, nil
}

func (svc *ResourcesService) Delete(_ context.Context, r *v1.ResourceDeleteRequest) (*v1.Resource, error) {
	res, err := db.DeleteResource(r.Id)
	if err != nil {
		return nil, err
	}

	messageMeta := db.MessageMeta{
		SentTimestamp:        0,
		ResourceGenerationID: res.ResourceGenerationID + 1,
	}

	resourceMessage := db.ResourceMessage{
		Id:          res.Id,
		ConsumerId:  res.ConsumerId,
		MessageMeta: messageMeta,
		Content:     &res.Object,
	}
	svc.resourceChan <- resourceMessage

	return &v1.Resource{Id: res.Id,
		ConsumerId:   res.ConsumerId,
		GenerationId: res.ResourceGenerationID}, nil
}