package memcached

import (
	"context"
	"reflect"

	cachev1alpha1 "github.com/swatishr/memcached-operator/pkg/apis/cache/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	
	appsv1 "k8s.io/api/apps/v1"
)

var log = logf.Log.WithName("controller_memcached")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Memcached Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMemcached{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("memcached-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Memcached
	err = c.Watch(&source.Kind{Type: &cachev1alpha1.Memcached{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Memcached
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cachev1alpha1.Memcached{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMemcached implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMemcached{}

// ReconcileMemcached reconciles a Memcached object
type ReconcileMemcached struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Memcached object and makes changes based on the state read
// and what is in the Memcached.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMemcached) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Memcached")

	// Fetch the Memcached instance
	instance := &cachev1alpha1.Memcached{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Memcached resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get Memcached")
		return reconcile.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	foundDep := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
	Name: instance.Name, Namespace: instance.Namespace,
	}, foundDep)

	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForMemcached(instance)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace",
				dep.Namespace, "Deployment.Name", dep.Name)
			r.updateStatus(instance, "Failed to create new Deployment")
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		r.updateStatus(instance, "Deployment created succesfully")
		return reconcile.Result{Requeue: true}, nil
	}

	// Ensure the deployment size is the same as the spec
	size := instance.Spec.NumOfReplicas
	if *foundDep.Spec.Replicas != size {
		foundDep.Spec.Replicas = &size
		err = r.client.Update(context.TODO(), foundDep)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace",
			foundDep.Namespace, "Deployment.Name", foundDep.Name)
			r.updateStatus(instance, "Reconcile failed: Failed to update Deployment")
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	r.updateStatus(instance, "Memcached deployed succesfully")
	return reconcile.Result{}, nil
}

// labelsForMemcached returns the labels for selecting the resources
// belonging to the given memcached CR name.
// used in the 'spec.selector.matchLabels' and 'spec.template.metadata.labels' fields of the deployment
func labelsForMemcached(name string) map[string]string {
    return map[string]string{"app": "memcached", "memcached_cr": name}
}

// deploymentForMemcached returns a memcached Deployment object
func (r *ReconcileMemcached) deploymentForMemcached(m *cachev1alpha1.Memcached) *appsv1.Deployment {
    ls := labelsForMemcached(m.Name)    // Build labels based on deployment name.
    replicas := m.Spec.NumOfReplicas    // Use the NumOfReplicas from the CRD Spec for the replica count

    dep := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      m.Name,
            Namespace: m.Namespace,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: ls,
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{{
                        Image:   "memcached:1.4.36-alpine",
                        Name:    "memcached",
                        Command: []string{"memcached", "-m=64", "-o", "modern", "-v"},
                        Ports: []corev1.ContainerPort{{
                            ContainerPort: 11211,
                            Name:          "memcached",
                        }},
                    }},
                },
            },
        },
	}
	// Set Memcached instance as the owner and controller
    controllerutil.SetControllerReference(m, dep, r.scheme)
    return dep
}

// updateStatus updates the Memcached CR's status field OverallStatus
func (r *ReconcileMemcached) updateStatus(m *cachev1alpha1.Memcached, message string) {
	if !reflect.DeepEqual(m.Status.OverallStatus, message) {
		m.Status.OverallStatus = message
		if err := r.client.Status().Update(context.TODO(), m); err != nil {
			log.Error(err, "Error updating instance status")
		}
	}
}
