/*
Copyright 2019 FinTech Studios, Inc.

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
	"time"

	"github.com/fintechstudios/ververica-platform-k8s-operator/api/v1beta1/converters"
	appManagerApi "github.com/fintechstudios/ververica-platform-k8s-operator/appmanager-api-client"
	"github.com/fintechstudios/ververica-platform-k8s-operator/controllers/annotations"
	appManager "github.com/fintechstudios/ververica-platform-k8s-operator/controllers/appmanager"
	"github.com/fintechstudios/ververica-platform-k8s-operator/controllers/utils"
	"github.com/go-logr/logr"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ververicaplatformv1beta1 "github.com/fintechstudios/ververica-platform-k8s-operator/api/v1beta1"
)

// VpDeploymentTargetReconciler reconciles a VpDeploymentTarget object
type VpDeploymentTargetReconciler struct {
	client.Client
	Log                 logr.Logger
	AppManagerAPIClient *appManagerApi.APIClient
	AppManagerAuthStore *appManager.AuthStore
}

// updateResource takes a k8s resource and a VP resource and merges them
func (r *VpDeploymentTargetReconciler) updateResource(resource *ververicaplatformv1beta1.VpDeploymentTarget, depTarget *appManagerApi.DeploymentTarget) error {
	ctx := context.Background()

	if resource.Annotations == nil {
		resource.Annotations = make(map[string]string)
	}

	annotations.Set(resource.Annotations,
		annotations.Pair(annotations.ID, depTarget.Metadata.Id),
		annotations.Pair(annotations.ResourceVersion, strconv.Itoa(int(depTarget.Metadata.ResourceVersion))))

	if err := r.Update(ctx, resource); err != nil {
		return err
	}

	return nil
}

// getLogger creates a logger for the controller with the request name
func (r *VpDeploymentTargetReconciler) getLogger(req ctrl.Request) logr.Logger {
	return r.Log.WithValues("vpdeploymenttarget", req.NamespacedName)
}

// handleCreate creates VP resources
func (r *VpDeploymentTargetReconciler) handleCreate(req ctrl.Request, vpDepTarget ververicaplatformv1beta1.VpDeploymentTarget) (ctrl.Result, error) {
	log := r.getLogger(req)
	nsName := utils.GetNamespaceOrDefault(vpDepTarget.Spec.Metadata.Namespace)
	ctx, err := r.AppManagerAuthStore.ContextForNamespace(context.Background(), nsName)
	if err != nil {
		log.Error(err, "cannot create context")
		return ctrl.Result{Requeue: false}, nil
	}

	patchSet, err := converters.DeploymentTargetPatchSetFromNative(vpDepTarget.Spec.Spec.DeploymentPatchSet)
	if err != nil {
		return ctrl.Result{}, err
	}

	depTarget := appManagerApi.DeploymentTarget{
		ApiVersion: "v1",
		Metadata: &appManagerApi.DeploymentTargetMetadata{
			Name:        req.Name,
			Namespace:   vpDepTarget.Spec.Metadata.Namespace,
			Labels:      vpDepTarget.Spec.Metadata.Labels,
			Annotations: vpDepTarget.Spec.Metadata.Annotations,
		},
		Spec: &appManagerApi.DeploymentTargetSpec{
			Kubernetes:         &appManagerApi.KubernetesTarget{Namespace: vpDepTarget.Spec.Spec.Kubernetes.Namespace},
			DeploymentPatchSet: patchSet,
		},
	}
	// create it
	createdDepTarget, res, err := r.AppManagerAPIClient.
		DeploymentTargetsApi.
		CreateDeploymentTarget(ctx, nsName, depTarget)

	if res != nil && res.StatusCode == 400 {
		// Bad request, don't requeue
		return ctrl.Result{Requeue: false}, err
	}

	if err != nil {
		log.Error(err, "Error creating VP Deployment Target")
		return ctrl.Result{}, err
	}

	log.Info("Created deployment target")

	// Now update the k8s resource and status as well
	if err := r.updateResource(&vpDepTarget, &createdDepTarget); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// handleUpdate updates the k8s resource when it already exists in the VP
// updates are not supported on Deployment Targets in the VP API, so just need to mirror the latest state
func (r *VpDeploymentTargetReconciler) handleUpdate(req ctrl.Request, vpDepTarget ververicaplatformv1beta1.VpDeploymentTarget, depTarget appManagerApi.DeploymentTarget) (ctrl.Result, error) {
	r.getLogger(req).Info("cannot update deployment targets in the Ververica Platform - must delete and recreate")
	err := r.updateResource(&vpDepTarget, &depTarget)
	return ctrl.Result{}, err
}

// handleDelete will ensure that the Ververica Platform namespace is also cleaned up
func (r *VpDeploymentTargetReconciler) handleDelete(req ctrl.Request, vpDepTarget ververicaplatformv1beta1.VpDeploymentTarget) (ctrl.Result, error) {
	log := r.getLogger(req)
	nsName := utils.GetNamespaceOrDefault(vpDepTarget.Spec.Metadata.Namespace)
	ctx, err := r.AppManagerAuthStore.ContextForNamespace(context.Background(), nsName)
	if err != nil {
		log.Error(err, "cannot create context")
		return ctrl.Result{Requeue: false}, nil
	}

	// Let's make sure it's deleted from the ververica platform
	depTarget, res, err := r.AppManagerAPIClient.DeploymentTargetsApi.DeleteDeploymentTarget(ctx, nsName, req.Name)

	if res != nil && res.StatusCode == 409 {
		// Conflict - still have deployments referenced
		// Can take a while to tear down
		return ctrl.Result{RequeueAfter: time.Second * 30}, nil
	}

	if err != nil {
		// If it's already gone, great!
		return ctrl.Result{}, utils.IgnoreNotFoundError(err)
	}

	log.Info("Deleting Deployment Target", "name", depTarget.Metadata.Name)
	return ctrl.Result{}, nil
}

// +kubebuilder:rbac:groups=ververicaplatform.fintechstudios.com,resources=vpdeploymenttargets,verbs=get;list;watch;create;update;patch;delete

// Reconcile tries to make the current state more like the desired state
func (r *VpDeploymentTargetReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.getLogger(req)

	var vpDepTarget ververicaplatformv1beta1.VpDeploymentTarget
	// If it's gone, it's gone!
	if err := r.Get(ctx, req.NamespacedName, &vpDepTarget); err != nil {
		return ctrl.Result{}, utils.IgnoreNotFoundError(err)
	}

	if vpDepTarget.ObjectMeta.DeletionTimestamp.IsZero() {
		// Not being deleted, add the finalizer
		if utils.AddFinalizer(&vpDepTarget.ObjectMeta) {
			log.Info("Adding Finalizer")
			if err := r.Update(ctx, &vpDepTarget); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// Being deleted
		log.Info("Deletion event", "name", req.Name)
		res, err := r.handleDelete(req, vpDepTarget)
		if utils.IsRequeueResponse(res, err) {
			return res, err
		}
		// otherwise, we're all good, just remove the finalizer
		if utils.RemoveFinalizer(&vpDepTarget.ObjectMeta) {
			if err := r.Update(ctx, &vpDepTarget); err != nil {
				return ctrl.Result{}, err
			}
		}

		return res, nil
	}

	nsName := utils.GetNamespaceOrDefault(vpDepTarget.Spec.Metadata.Namespace)
	appManagerCtx, err := r.AppManagerAuthStore.ContextForNamespace(context.Background(), nsName)
	if err != nil {
		log.Error(err, "cannot create context")
		return ctrl.Result{Requeue: false}, nil
	}
	depTarget, _, err := r.AppManagerAPIClient.DeploymentTargetsApi.GetDeploymentTarget(appManagerCtx, nsName, req.Name)
	if err != nil {
		if utils.IsNotFoundError(err) {
			// Not found, let's create it
			return r.handleCreate(req, vpDepTarget)
		}
		// Other error, not good!
		return ctrl.Result{}, err
	}

	log.Info("Update event")
	return r.handleUpdate(req, vpDepTarget, depTarget)
}

// SetupWithManager is a helper function to initial on manager boot
func (r *VpDeploymentTargetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ververicaplatformv1beta1.VpDeploymentTarget{}).
		Complete(r)
}
