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
	"context"
	"encoding/json"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	velerov1api "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/persistence"
	"github.com/vmware-tanzu/velero/pkg/plugin/clientmgmt"
	"github.com/vmware-tanzu/velero/pkg/util/kube"
)

const signedURLTTL = 10 * time.Minute

type DownloadRequest struct {
	// use variables to refer to these functions so they can be
	// replaced with fakes for testing.
	NewPluginManager func(logrus.FieldLogger) clientmgmt.Manager
	NewBackupStore   func(*velerov1api.BackupStorageLocation, persistence.ObjectStoreGetter, logrus.FieldLogger) (persistence.BackupStore, error)
}

// generatePreSignedURL generates a pre-signed URL for downloadRequest, changes the phase to
// Processed, and persists the changes to storage.
func (d *DownloadRequest) generatePreSignedURL(downloadRequest *velerov1api.DownloadRequest, log logrus.FieldLogger) error {
	update := downloadRequest.DeepCopy()

	var (
		backupName string
		err        error
	)

	switch downloadRequest.Spec.Target.Kind {
	case velerov1api.DownloadTargetKindRestoreLog, velerov1api.DownloadTargetKindRestoreResults:
		restore, err := c.restoreLister.Restores(downloadRequest.Namespace).Get(downloadRequest.Spec.Target.Name)
		if err != nil {
			return errors.Wrap(err, "error getting Restore")
		}

		backupName = restore.Spec.BackupName
	default:
		backupName = downloadRequest.Spec.Target.Name
	}

	backup, err := c.backupLister.Backups(downloadRequest.Namespace).Get(backupName)
	if err != nil {
		return errors.WithStack(err)
	}

	backupLocation := &velerov1api.BackupStorageLocation{}
	if err := c.kbClient.Get(context.Background(), client.ObjectKey{
		Namespace: backup.Namespace,
		Name:      backup.Spec.StorageLocation,
	}, backupLocation); err != nil {
		return errors.WithStack(err)
	}

	pluginManager := c.newPluginManager(log)
	defer pluginManager.CleanupClients()

	backupStore, err := c.newBackupStore(backupLocation, pluginManager, log)
	if err != nil {
		return errors.WithStack(err)
	}

	if update.Status.DownloadURL, err = backupStore.GetDownloadURL(downloadRequest.Spec.Target); err != nil {
		return err
	}

	update.Status.Phase = velerov1api.DownloadRequestPhaseProcessed
	update.Status.Expiration = &metav1.Time{Time: c.clock.Now().Add(persistence.DownloadURLTTL)}

	_, err = patchDownloadRequest(downloadRequest, update, c.downloadRequestClient)
	return errors.WithStack(err)
}

// DeleteIfExpired deletes downloadRequest if it has expired.
func (d *DownloadRequest) DeleteIfExpired(downloadRequest *velerov1api.DownloadRequest) error {
	log := c.logger.WithField("key", kube.NamespaceAndName(downloadRequest))
	log.Info("checking for expiration of DownloadRequest")
	if downloadRequest.Status.Expiration.Time.After(c.clock.Now()) {
		log.Debug("DownloadRequest has not expired")
		return nil
	}

	log.Debug("DownloadRequest has expired - deleting")
	return errors.WithStack(c.downloadRequestClient.DownloadRequests(downloadRequest.Namespace).Delete(context.TODO(), downloadRequest.Name, metav1.DeleteOptions{}))
}

// resync requeues all the DownloadRequests in the lister's cache. This is mostly to handle deleting
// any expired requests that were not deleted as part of the normal client flow for whatever reason.
// func (d *DownloadRequest) resync() {
// 	list, err := c.downloadRequestLister.List(labels.Everything())
// 	if err != nil {
// 		c.logger.WithError(errors.WithStack(err)).Error("error listing download requests")
// 		return
// 	}

// 	for _, dr := range list {
// 		key, err := cache.MetaNamespaceKeyFunc(dr)
// 		if err != nil {
// 			c.logger.WithError(errors.WithStack(err)).WithField("downloadRequest", dr.Name).Error("error generating key for download request")
// 			continue
// 		}

// 		c.queue.Add(key)
// 	}
// }

func patchDownloadRequest(original, updated *velerov1api.DownloadRequest, client velerov1client.DownloadRequestsGetter) (*velerov1api.DownloadRequest, error) {
	origBytes, err := json.Marshal(original)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling original download request")
	}

	updatedBytes, err := json.Marshal(updated)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling updated download request")
	}

	patchBytes, err := jsonpatch.CreateMergePatch(origBytes, updatedBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error creating json merge patch for download request")
	}

	res, err := client.DownloadRequests(original.Namespace).Patch(context.TODO(), original.Name, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error patching download request")
	}

	return res, nil
}
