/*

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

package pod_restarter

import (
	"context"
	"fmt"

	"github.com/openshift/cluster-resource-override-admission/pkg/api"
	"github.com/openshift/cluster-resource-override-admission/pkg/clusterresourceoverride"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	LabelOverridePodResources = fmt.Sprintf("%s.%s/enabled", clusterresourceoverride.Resource, api.Group)
	AnnotationPodScaled       = fmt.Sprintf("%s.%s/scaled", clusterresourceoverride.Resource, api.Group)
)

// PodRestartReconciler reconciles a namespaces objects
type PodRestartReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// SetupWithManager creates a new PerformanceProfile Controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func (r *PodRestartReconciler) SetupWithManager(mgr ctrl.Manager) error {
	namespacesPredicates := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Meta == nil {
				klog.Error("Create event has no metadata")
				return false
			}

			if e.Object == nil {
				klog.Error("Create event has no object")
				return false
			}

			namespace := e.Object.(*corev1.Namespace)
			return isNamespaceLabeled(namespace)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.MetaOld == nil {
				klog.Error("Update event has no old metadata")
				return false
			}
			if e.MetaNew == nil {
				klog.Error("Update event has no new metadata")
				return false
			}
			if e.ObjectOld == nil {
				klog.Error("Update event has no old runtime object to update")
				return false
			}
			if e.ObjectNew == nil {
				klog.Error("Update event has no new runtime object for update")
				return false
			}

			oldNamespace := e.ObjectOld.(*corev1.Namespace)
			newNamespace := e.ObjectNew.(*corev1.Namespace)

			// TODO: I am assuming that once the label was remove from the namespace we should re-scale pod
			// should re-check if it needed at all
			return isNamespaceLabeled(oldNamespace) != isNamespaceLabeled(newNamespace)
		},
	}
	err := ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}, builder.WithPredicates(namespacesPredicates)).
		Complete(r)

	return err
}

func isNamespaceLabeled(namespace *corev1.Namespace) bool {
	if v, ok := namespace.Labels[LabelOverridePodResources]; !ok || v != "true" {
		return false
	}

	return true
}

func (r *PodRestartReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	// Fetch the namespace
	namespace := &corev1.Namespace{}
	err := r.Get(context.TODO(), req.NamespacedName, namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// scale pods CPU requests
	if err := r.scalePodsCPURequests(namespace); err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PodRestartReconciler) scalePodsCPURequests(namespace *corev1.Namespace) error {
	// List all pods placed under the namespace
	pods := &corev1.PodList{}
	if err := r.List(context.TODO(), pods, client.InNamespace(namespace.Name)); err != nil {
		return err
	}

	for i, pod := range pods.Items {
		if isPodScaled(&pods.Items[i]) {
			continue
		}

		// we do not want to delete pods that already was deleted
		if pod.DeletionTimestamp != nil {
			continue
		}

		// do not delete pods in succeeded, failed or unknown phase
		if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodPending {
			continue
		}

		if err := r.Delete(context.TODO(), &pods.Items[i]); err != nil {
			klog.Errorf("failed to delete pod %q", pod.Name)
			return err
		}
	}

	return nil
}

func isPodScaled(pod *corev1.Pod) bool {
	if v, ok := pod.Annotations[AnnotationPodScaled]; !ok || v != "true" {
		return false
	}

	return true
}
