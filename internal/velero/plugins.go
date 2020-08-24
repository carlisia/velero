/*
Copyright 2020 the Velero contributors.

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

package velero

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	velerov1api "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/persistence"
	"github.com/vmware-tanzu/velero/pkg/plugin/clientmgmt"
)

// BackupStoreManager holds the components necessary to fetch a backup store
type BackupStoreManager struct {
	// use variables to refer to these functions so they can be
	// replaced with fakes for testing.
	newPluginManager func(logrus.FieldLogger) clientmgmt.Manager
	newBackupStore   func(*velerov1api.BackupStorageLocation, persistence.ObjectStoreGetter, logrus.FieldLogger) (persistence.BackupStore, error)
}

// NewBackupStoreManager  returns a BackupStoreManager
func NewBackupStoreManager(newPluginManager func(logrus.FieldLogger) clientmgmt.Manager,
	newBackupStore func(*velerov1api.BackupStorageLocation, persistence.ObjectStoreGetter, logrus.FieldLogger) (persistence.BackupStore, error)) BackupStoreManager {

	return BackupStoreManager{newPluginManager, newBackupStore}
}

//GetBackupStore returns an initialized backup store
func (b BackupStoreManager) GetBackupStore(backupLocation *velerov1api.BackupStorageLocation, log logrus.FieldLogger) (persistence.BackupStore, error) {
	pluginManager := b.newPluginManager(log)
	defer pluginManager.CleanupClients()

	backupStore, err := b.newBackupStore(backupLocation, pluginManager, log)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return backupStore, nil
}
