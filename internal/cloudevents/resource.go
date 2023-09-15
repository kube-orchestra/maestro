package cloudevents

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/kube-orchestra/maestro/internal/db"
	cegeneric "open-cluster-management.io/api/cloudevents/generic"
	cetypes "open-cluster-management.io/api/cloudevents/generic/types"
)

type ResourceLister struct{}

var _ cegeneric.Lister[*db.Resource] = &ResourceLister{}

func (resLister *ResourceLister) List(listOpts cetypes.ListOptions) ([]*db.Resource, error) {
	resources, err := db.ListResourceByConsumer(listOpts.ClusterName)
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func ResourceStatusHashGetter(obj *db.Resource) (string, error) {
	statusBytes, err := json.Marshal(obj.Status)
	if err != nil {
		return "", fmt.Errorf("failed to marshal work status, %v", err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(statusBytes)), nil
}