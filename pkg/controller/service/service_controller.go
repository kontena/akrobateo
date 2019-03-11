package service

import (
	"context"
	"crypto/md5"
	"fmt"
	"reflect"
	"sort"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("service-lb-controller")

var (
	trueVal = true
)

const (
	lbImage           = "docker.io/rancher/klipper-lb:v0.1.1"
	svcNameLabel      = "servicelb.kontena.io/svcname"
	svcHashAnnotation = "servicelb.kontena.io/svchash"
)

// Add creates a new Service Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("service-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource and requeue the owner Service
	// Changes in secondary resource(s) trigger reconcile for the original (owner) service
	// From that event we can go and update the service IPs etc.
	err = c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &corev1.Service{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileService{}

// ReconcileService reconciles a Service object
type ReconcileService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Service object and makes changes based on the state read
// and what is in the Service.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Service")

	// Fetch the Service instance
	svc := &corev1.Service{}
	err := r.client.Get(context.TODO(), request.NamespacedName, svc)
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

	if svc.Spec.Type != corev1.ServiceTypeLoadBalancer || svc.Spec.ClusterIP == "" || svc.Spec.ClusterIP == "None" {
		// WE're only interested in LoadBalancer type services
		// Return and don't requeue
		reqLogger.Info("Not a LoadBalancer type of service, ignoring")
		return reconcile.Result{}, nil
	}

	// Generate the needed DS for the service
	ds := newDaemonSetForService(svc)

	// Set Service instance as the owner and controller
	if err := controllerutil.SetControllerReference(svc, ds, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if the needed DaemonSet already exists
	found := &appsv1.DaemonSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: ds.Name, Namespace: ds.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new DS", "DS.Namespace", ds.Namespace, "DS.Name", ds.Name)
		err = r.client.Create(context.TODO(), ds)
		if err != nil {
			return reconcile.Result{}, err
		}

		// DaemonSet created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if found.Annotations[svcHashAnnotation] != serviceHash(svc) {
		// Need to update the DS
		reqLogger.Info("Updating DS", "DS.Namespace", ds.Namespace, "DS.Name", ds.Name)
		err = r.client.Update(context.TODO(), ds)
		if err != nil {
			return reconcile.Result{}, err
		}
		// DaemonSet updated successfully - don't requeue
		// Changes in the DS will trigger proper requests for this as the DS rollout progresses
		return reconcile.Result{}, nil
	}

	// DS already exists - don't requeue but sync up the addresses for the service
	// We also get the reconcile request on changes to the created DS, so for each of those try to sync the service addresses
	err = r.syncServiceAddresses(svc)
	if err != nil {
		reqLogger.Error(err, "Failed to sync service addresses")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func getDSName(svc *corev1.Service) string {
	return fmt.Sprintf("service-lb-%s", svc.Name)
}

func newDaemonSetForService(svc *corev1.Service) *appsv1.DaemonSet {
	name := getDSName(svc)

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: svc.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					Name:       svc.Name,
					APIVersion: "v1",
					Kind:       "Service",
					UID:        svc.UID,
					Controller: &trueVal,
				},
			},
			Annotations: map[string]string{
				// Used to trigger DS update if the service spec changes
				svcHashAnnotation: serviceHash(svc),
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":        name,
						svcNameLabel: svc.Name,
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						corev1.Container{
							Name:  "sysctl",
							Image: lbImage,
							Command: []string{
								"sh",
								"-c",
								"sysctl -w net.ipv4.ip_forward=1",
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &trueVal,
							},
						},
					},
				},
			},
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.RollingUpdateDaemonSetStrategyType,
			},
		},
	}

	// Create a container per exposed port
	for i, port := range svc.Spec.Ports {
		portName := port.Name
		if portName == "" {
			portName = fmt.Sprintf("port-%d", i)
		}
		container := corev1.Container{
			Name:            portName,
			Image:           lbImage,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Ports: []corev1.ContainerPort{
				{
					Name:          portName,
					ContainerPort: port.Port,
					HostPort:      port.Port,
				},
			},
			Env: []corev1.EnvVar{
				{
					Name:  "SRC_PORT",
					Value: strconv.Itoa(int(port.Port)),
				},
				{
					Name:  "DEST_PROTO",
					Value: string(port.Protocol),
				},
				{
					Name:  "DEST_PORT",
					Value: port.TargetPort.String(),
				},
				{
					Name:  "DEST_IP",
					Value: svc.Spec.ClusterIP,
				},
			},
			SecurityContext: &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{
						"NET_ADMIN",
					},
				},
			},
		}

		ds.Spec.Template.Spec.Containers = append(ds.Spec.Template.Spec.Containers, container)
	}

	return ds
}

func (r *ReconcileService) syncServiceAddresses(svc *corev1.Service) error {
	sw := ServiceWrangler{
		client:  r.client,
		service: *svc,
	}

	pods, err := sw.FindPods()
	if err != nil {
		return err
	}

	ips, err := r.podIPs(pods.Items)
	if err != nil {
		return err
	}

	existingIPs := sw.ExistingIPs()

	sort.Strings(ips)
	sort.Strings(existingIPs)

	if reflect.DeepEqual(ips, existingIPs) {
		log.Info("Existing service addresses match, no need to update")
		return nil
	}

	log.Info("Addresses need to be updated for service:", "IPs", ips)

	return sw.UpdateAddresses(ips)
}

func (r *ReconcileService) getNode(name string) (*corev1.Node, error) {
	node := &corev1.Node{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ""}, node); err != nil {
		return nil, err
	}
	return node, nil
}

// Calculates the "checksum" for the services spec part.
// This is used to track update need for the created daemonset.
func serviceHash(svc *corev1.Service) string {
	d, err := svc.Spec.Marshal()
	if err != nil {
		return ""
	}

	checksum := fmt.Sprintf("%x", md5.Sum(d))

	return checksum
}
