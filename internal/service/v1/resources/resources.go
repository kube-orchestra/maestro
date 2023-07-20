package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/kube-orchestra/maestro/db"
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
}

func NewResourceService() *ResourcesService {
	return &ResourcesService{}
}

func (svc *ResourcesService) Read(_ context.Context, r *v1.ResourceReadRequest) (*v1.Resource, error) {
	res, err := db.GetResource(r.Id)
	if err != nil {
		return nil, err
	}

	fmt.Println(res.Object)

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
	res := db.Resource{
		Id:         uuid.NewString(),
		ConsumerId: r.ConsumerId,
		Object:     unstructured.Unstructured{Object: r.Object.AsMap()},
	}

	err := db.PutResource(&res)
	if err != nil {
		return nil, err
	}

	return &v1.Resource{Id: res.Id, ConsumerId: res.ConsumerId, Object: r.Object}, nil
}
