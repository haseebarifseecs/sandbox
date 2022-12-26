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
	"strconv"

	stakatoriov1alpha1 "github.com/haseebarifseecs/sandbox/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// UserReconciler reconciles a User object
type UserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=stakator.io.stakator.io,resources=users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=stakator.io.stakator.io,resources=users/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=stakator.io.stakator.io,resources=users/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the User object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Inside Reconciler")
	user := &stakatoriov1alpha1.User{}
	err := r.Get(ctx, req.NamespacedName, user)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("User CRD Doesn't exist or it has been deleted")
			return ctrl.Result{}, nil
		}

		log.Info("Some error has occured", err)
		return ctrl.Result{}, err
	}
	count := user.Spec.SandboxCount
	username := user.Spec.Name
	sandboxName := "SB-" + username + "-" + strconv.Itoa(count)
	found := &stakatoriov1alpha1.Sandbox{}
	err = r.Get(ctx, types.NamespacedName{Name: user.Name, Namespace: user.Namespace}, found)
	if err != nil {
		if apierrors.IsNotFound(err) {
			for i := 0; i < count; i++ {
				err = r.Create(ctx, &stakatoriov1alpha1.Sandbox{
					ObjectMeta: metav1.ObjectMeta{
						Name:      user.Name,
						Namespace: user.Namespace,
					},
					Spec: stakatoriov1alpha1.SandboxSpec{
						Name: sandboxName,
						Type: "T1",
					},
				})
				if err != nil {
					log.Error(err, "Error")
					// log.Info("Failed to create Sandbox Resource", err)
					return ctrl.Result{}, err
				}
			}
		}
	}

	log.Info("HELLO \n")
	log.Info("SandBox Count \t" + strconv.Itoa(user.Status.SandboxCount))
	if user.Status.SandboxCount == 0 {
		log.Info("Inside Condition Met:")
		user.Status.SandboxCount = 1
		err = r.Status().Update(ctx, user)
		if err != nil {
			log.Info("Failed to update", err)
			return ctrl.Result{}, err
		}
	}
	log.Info("Sandbox Count \t" + strconv.Itoa(user.Status.SandboxCount))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&stakatoriov1alpha1.User{}).
		Complete(r)
}
