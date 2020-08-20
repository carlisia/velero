/*
Copyright 2017 the Velero contributors.

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

package controller

import (
	"context"

	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	kbclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/velero/internal/velero"
	velerov1api "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

// DownloadRequestReconciler reconciles a BackupStorageLocation object
type DownloadRequestReconciler struct {
	Scheme          *runtime.Scheme
	Client          kbclient.Client
	Ctx             context.Context
	DownloadRequest velero.DownloadRequest

	Log logrus.FieldLogger
}

// +kubebuilder:rbac:groups=velero.io,resources=backupstoragelocations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=velero.io,resources=downloadrequests/status,verbs=get;update;patch
func (r *DownloadRequestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithField("controller", "download-request")

	r.Log.Debug("Running processDownloadRequest")
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.WithError(err).Error("error splitting queue key")
		return ctrl.Result{Requeue: true}, err
	}

	downloadRequest, err := c.downloadRequestLister.DownloadRequests(ns).Get(name)
	if apierrors.IsNotFound(err) {
		log.Debug("Unable to find DownloadRequest")
		return ctrl.Result{Requeue: true}, err
	}
	if err != nil {
		// return errors.Wrap(err, "error getting DownloadRequest")
		return ctrl.Result{Requeue: true}, err
	}

	switch downloadRequest.Status.Phase {
	case "", velerov1api.DownloadRequestPhaseNew:
		return c.generatePreSignedURL(downloadRequest, log)
	case velerov1api.DownloadRequestPhaseProcessed:
		return c.DeleteIfExpired(downloadRequest)
	}

	// Requeue is mostly to handle deleting any expired requests that were not
	// deleted as part of the normal client flow for whatever reason.
	return ctrl.Result{Requeue: true}, nil
}

func (r *DownloadRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&velerov1api.BackupStorageLocation{}).
		Complete(r)
}

// type downloadRequestController struct {
// 	*genericController

// 	downloadRequestClient velerov1client.DownloadRequestsGetter
// 	downloadRequestLister velerov1listers.DownloadRequestLister
// 	restoreLister         velerov1listers.RestoreLister
// 	clock                 clock.Clock
// 	kbClient              client.Client
// 	backupLister          velerov1listers.BackupLister
// 	newPluginManager      func(logrus.FieldLogger) clientmgmt.Manager
// 	newBackupStore        func(*velerov1api.BackupStorageLocation, persistence.ObjectStoreGetter, logrus.FieldLogger) (persistence.BackupStore, error)
// }

// // NewDownloadRequestController creates a new DownloadRequestController.
// func NewDownloadRequestController(
// 	downloadRequestClient velerov1client.DownloadRequestsGetter,
// 	downloadRequestInformer velerov1informers.DownloadRequestInformer,
// 	restoreLister velerov1listers.RestoreLister,
// 	kbClient client.Client,
// 	backupLister velerov1listers.BackupLister,
// 	newPluginManager func(logrus.FieldLogger) clientmgmt.Manager,
// 	logger logrus.FieldLogger,
// ) Interface {
// 	c := &downloadRequestController{
// 		genericController:     newGenericController("downloadrequest", logger),
// 		downloadRequestClient: downloadRequestClient,
// 		downloadRequestLister: downloadRequestInformer.Lister(),
// 		restoreLister:         restoreLister,
// 		kbClient:              kbClient,
// 		backupLister:          backupLister,

// 		// use variables to refer to these functions so they can be
// 		// replaced with fakes for testing.
// 		newPluginManager: newPluginManager,
// 		newBackupStore:   persistence.NewObjectBackupStore,

// 		clock: &clock.RealClock{},
// 	}

// 	c.syncHandler = c.processDownloadRequest

// 	downloadRequestInformer.Informer().AddEventHandler(
// 		cache.ResourceEventHandlerFuncs{
// 			AddFunc: func(obj interface{}) {
// 				key, err := cache.MetaNamespaceKeyFunc(obj)
// 				if err != nil {
// 					downloadRequest := obj.(*velerov1api.DownloadRequest)
// 					c.logger.WithError(errors.WithStack(err)).
// 						WithField("downloadRequest", downloadRequest.Name).
// 						Error("Error creating queue key, item not added to queue")
// 					return
// 				}
// 				c.queue.Add(key)
// 			},
// 		},
// 	)

// 	return c
// }
