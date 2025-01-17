package v1beta3

import (
	"context"
	"fmt"
	k8sCoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	utilRuntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sWatchV1 "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	tarsCrdV1beta3 "k8s.tars.io/crd/v1beta3"
	tarsMetaV1beta3 "k8s.tars.io/meta/v1beta3"
	"strings"
	"tarscontroller/controller"
	"tarscontroller/reconcile"
	"time"
)

type PVCReconciler struct {
	clients   *controller.Clients
	informers *controller.Informers
	threads   int
	workQueue workqueue.RateLimitingInterface
}

func NewPVCReconciler(clients *controller.Clients, informers *controller.Informers, threads int) *PVCReconciler {
	reconciler := &PVCReconciler{
		clients:   clients,
		informers: informers,
		threads:   threads,
		workQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), ""),
	}
	informers.Register(reconciler)
	return reconciler
}

func (r *PVCReconciler) processItem() bool {

	obj, shutdown := r.workQueue.Get()

	if shutdown {
		return false
	}

	defer r.workQueue.Done(obj)

	key, ok := obj.(string)
	if !ok {
		utilRuntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
		r.workQueue.Forget(obj)
		return true
	}

	res := r.reconcile(key)

	switch res {
	case reconcile.AllOk:
		r.workQueue.Forget(obj)
		return true
	case reconcile.RateLimit:
		r.workQueue.AddRateLimited(obj)
		return true
	case reconcile.FatalError:
		r.workQueue.ShutDown()
		return false
	case reconcile.AddAfter:
		r.workQueue.AddAfter(obj, time.Second*3)
		return true
	default:
		//code should not reach here
		utilRuntime.HandleError(fmt.Errorf("should not reach place"))
		return false
	}
}

func (r *PVCReconciler) EnqueueObj(resourceName string, resourceEvent k8sWatchV1.EventType, resourceObj interface{}) {
	switch resourceObj.(type) {
	case *tarsCrdV1beta3.TServer:
		tserver := resourceObj.(*tarsCrdV1beta3.TServer)
		key := fmt.Sprintf("%s/%s", tserver.Namespace, tserver.Name)
		r.workQueue.Add(key)
	case *k8sCoreV1.PersistentVolumeClaim:
		if resourceEvent == k8sWatchV1.Deleted {
			break
		}
		pvc := resourceObj.(*k8sCoreV1.PersistentVolumeClaim)
		if pvc.Labels != nil {
			app, appExist := pvc.Labels[tarsMetaV1beta3.TServerAppLabel]
			server, serverExist := pvc.Labels[tarsMetaV1beta3.TServerNameLabel]
			if appExist && serverExist {
				key := fmt.Sprintf("%s/%s-%s", pvc.Namespace, strings.ToLower(app), strings.ToLower(server))
				r.workQueue.Add(key)
				return
			}
		}
	}
}

func (r *PVCReconciler) Start(stopCh chan struct{}) {
	for i := 0; i < r.threads; i++ {
		workFun := func() {
			for r.processItem() {
			}
			r.workQueue.ShutDown()
		}
		go wait.Until(workFun, time.Second, stopCh)
	}
}

func buildPVCAnnotations(tserver *tarsCrdV1beta3.TServer) map[string]map[string]string {
	var annotations = make(map[string]map[string]string, 0)
	if tserver.Spec.K8S.Mounts != nil {
		for _, mount := range tserver.Spec.K8S.Mounts {
			if mount.Source.TLocalVolume != nil {
				annotations[mount.Name] = map[string]string{
					tarsMetaV1beta3.TLocalVolumeUIDLabel:  mount.Source.TLocalVolume.UID,
					tarsMetaV1beta3.TLocalVolumeGIDLabel:  mount.Source.TLocalVolume.GID,
					tarsMetaV1beta3.TLocalVolumeModeLabel: mount.Source.TLocalVolume.Mode,
				}
			}
		}
	}
	return annotations
}

func (r *PVCReconciler) reconcile(key string) reconcile.Result {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilRuntime.HandleError(fmt.Errorf("invalid key: %s", key))
		return reconcile.AllOk
	}

	tserver, err := r.informers.TServerInformer.Lister().TServers(namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utilRuntime.HandleError(fmt.Errorf(tarsMetaV1beta3.ResourceGetError, "tserver", namespace, name, err.Error()))
			return reconcile.RateLimit
		}
		return reconcile.AllOk
	}

	if tserver.DeletionTimestamp != nil || tserver.Spec.K8S.DaemonSet {
		return reconcile.AllOk
	}

	annotations := buildPVCAnnotations(tserver)

	for volumeName, volumeProperties := range annotations {
		appRequirement, _ := labels.NewRequirement(tarsMetaV1beta3.TServerAppLabel, selection.DoubleEquals, []string{tserver.Spec.App})
		serverRequirement, _ := labels.NewRequirement(tarsMetaV1beta3.TServerNameLabel, selection.DoubleEquals, []string{tserver.Spec.Server})
		localVolumeRequirement, _ := labels.NewRequirement(tarsMetaV1beta3.TLocalVolumeLabel, selection.DoubleEquals, []string{volumeName})
		labelSelector := labels.NewSelector().Add(*appRequirement, *serverRequirement, *localVolumeRequirement)
		pvcs, err := r.informers.PersistentVolumeClaimInformer.Lister().PersistentVolumeClaims(namespace).List(labelSelector)
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			utilRuntime.HandleError(fmt.Errorf(tarsMetaV1beta3.ResourceSelectorError, namespace, "persistentVolumeclaims", err.Error()))
			return reconcile.RateLimit
		}

		for _, pvc := range pvcs {
			if pvc.DeletionTimestamp != nil || ContainLabel(pvc.Annotations, volumeProperties) {
				continue
			}
			pvcCopy := pvc.DeepCopy()
			if pvcCopy.Annotations != nil {
				for k, v := range volumeProperties {
					pvcCopy.Annotations[k] = v
				}
			} else {
				pvcCopy.Annotations = volumeProperties
			}
			_, err := r.clients.K8sClient.CoreV1().PersistentVolumeClaims(namespace).Update(context.TODO(), pvcCopy, k8sMetaV1.UpdateOptions{})
			if err == nil {
				continue
			}
			utilRuntime.HandleError(fmt.Errorf(tarsMetaV1beta3.ResourceUpdateError, "PersistentVolumeClaims", namespace, pvc.Name, err.Error()))
			return reconcile.RateLimit
		}
	}
	return reconcile.AllOk
}
