/*
Copyright 2017, 2019 the Velero contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"encoding/json"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	api "github.com/heptio/velero/pkg/apis/velero/v1"
	proto "github.com/heptio/velero/pkg/plugin/generated"
	"github.com/heptio/velero/pkg/plugin/velero"
)

// NewBackupItemActionPlugin constructs a BackupItemActionPlugin.
func NewBackupItemActionPlugin(options ...PluginOption) *BackupItemActionPlugin {
	return &BackupItemActionPlugin{
		pluginBase: newPluginBase(options...),
	}
}

// BackupItemActionGRPCClient implements the backup/ItemAction interface and uses a
// gRPC client to make calls to the plugin server.
type BackupItemActionGRPCClient struct {
	*clientBase
	grpcClient proto.BackupItemActionClient
}

func newBackupItemActionGRPCClient(base *clientBase, clientConn *grpc.ClientConn) interface{} {
	return &BackupItemActionGRPCClient{
		clientBase: base,
		grpcClient: proto.NewBackupItemActionClient(clientConn),
	}
}

func (c *BackupItemActionGRPCClient) AppliesTo() (velero.ResourceSelector, error) {
	req := &proto.BackupItemActionAppliesToRequest{
		Plugin: c.plugin,
	}

	res, err := c.grpcClient.AppliesTo(context.Background(), req)
	if err != nil {
		return velero.ResourceSelector{}, fromGRPCError(err)
	}

	if res.ResourceSelector == nil {
		return velero.ResourceSelector{}, errors.New("no ResourceSelector present")
	}

	return velero.ResourceSelector{
		IncludedNamespaces: res.ResourceSelector.IncludedNamespaces,
		ExcludedNamespaces: res.ResourceSelector.ExcludedNamespaces,
		IncludedResources:  res.ResourceSelector.IncludedResources,
		ExcludedResources:  res.ResourceSelector.ExcludedResources,
		LabelSelector:      res.ResourceSelector.Selector,
	}, nil
}

func (c *BackupItemActionGRPCClient) Execute(item runtime.Unstructured, backup *api.Backup) (runtime.Unstructured, []velero.ResourceIdentifier, error) {
	itemJSON, err := json.Marshal(item.UnstructuredContent())
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	backupJSON, err := json.Marshal(backup)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	req := &proto.ExecuteRequest{
		Plugin: c.plugin,
		Item:   itemJSON,
		Backup: backupJSON,
	}

	res, err := c.grpcClient.Execute(context.Background(), req)
	if err != nil {
		return nil, nil, fromGRPCError(err)
	}

	var updatedItem unstructured.Unstructured
	if err := json.Unmarshal(res.Item, &updatedItem); err != nil {
		return nil, nil, errors.WithStack(err)
	}

	var additionalItems []velero.ResourceIdentifier

	for _, itm := range res.AdditionalItems {
		newItem := velero.ResourceIdentifier{
			GroupResource: schema.GroupResource{
				Group:    itm.Group,
				Resource: itm.Resource,
			},
			Namespace: itm.Namespace,
			Name:      itm.Name,
		}

		additionalItems = append(additionalItems, newItem)
	}

	return &updatedItem, additionalItems, nil
}
