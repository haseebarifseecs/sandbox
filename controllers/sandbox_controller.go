/*
Copyright 2022.

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

package controllers

import (
	"context"
	"strings"
	"time"

	stakatoriov1alpha1 "github.com/haseebarifseecs/sandbox/api/v1alpha1"
	apicorev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// SandboxReconciler reconciles a Sandbox object
type SandboxReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const sandboxFinalizer = "stakator.io.stakator.io/finalizer"

//+kubebuilder:rbac:groups=stakator.io.stakator.io,resources=sandboxes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=stakator.io.stakator.io,resources=sandboxes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=stakator.io.stakator.io,resources=sandboxes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Sandbox object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *SandboxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)
	sandbox := &stakatoriov1alpha1.Sandbox{}
	err := r.Get(ctx, req.NamespacedName, sandbox)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("CR Doesn't Exist")
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}
	if !controllerutil.ContainsFinalizer(sandbox, sandboxFinalizer) {
		log.Info("Adding Finalizer for Memcached")
		if ok := controllerutil.AddFinalizer(sandbox, sandboxFinalizer); !ok {
			log.Error(err, "Failed to add finalizer into the custom resource")
			return ctrl.Result{Requeue: true}, nil
		}

		if err = r.Update(ctx, sandbox); err != nil {
			log.Error(err, "Failed to update custom resource to add finalizer")
			return ctrl.Result{}, err
		}
	}

	if sandbox.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(sandbox, sandboxFinalizer) {
			log.Info("Received Deletion Timestamp")
			namespaceName := strings.ToLower(sandbox.Spec.Name)
			namespace := &apicorev1.Namespace{}
			err = r.Get(ctx, types.NamespacedName{Name: namespaceName}, namespace)
			if !apierrors.IsNotFound(err) {
				err = r.Delete(ctx, namespace)
				if err != nil {
					log.Info("Error deleting Namespace")
				} else {
					ok := controllerutil.RemoveFinalizer(sandbox, sandboxFinalizer)
					if !ok {
						log.Info("Error removing finalizer")
					}
				}
			} else {
				ok := controllerutil.RemoveFinalizer(sandbox, sandboxFinalizer)
				if !ok {
					log.Info("Error removing finalizer")
				}
			}
			_ = r.Update(ctx, sandbox)
			return ctrl.Result{}, nil

		}
	}
	name := sandbox.Spec.Name
	name = strings.ToLower(name)
	log.Info(name)
	if name != "" {
		namespace := &apicorev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
		// if err != nil {
		// 	return ctrl.Result{}, nil
		// }
		err = r.Create(ctx, namespace)

		if apierrors.IsAlreadyExists(err) {
			log.Info("Namespace Exists for the user")
			// _ = r.Get(ctx, types.NamespacedName{Name: name}, namespace)
			// _ = ctrl.SetControllerReference(sandbox, namespace, r.Scheme)
			// _ = r.Update(ctx, namespace)
			return ctrl.Result{RequeueAfter: time.Second * 5, Requeue: true}, nil
		}
		if err != nil {
			return ctrl.Result{}, err
		}

	}

	return ctrl.Result{}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *SandboxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&stakatoriov1alpha1.Sandbox{}).
		Owns(&apicorev1.Namespace{}).
		Complete(r)
}
