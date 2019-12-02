package eventlogger

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"time"

	eventloggerv1 "github.com/bakito/k8s-event-logger-operator/pkg/apis/eventlogger/v1"
	c "github.com/bakito/k8s-event-logger-operator/pkg/constants"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/yaml"
)

const (
	elConfigFileName    = "event-listener.conf"
	elAbsConfigDirPath  = "/opt/go/config"
	elAbsConfigFilePath = elAbsConfigDirPath + "/" + elConfigFileName
)

var (
	log                    = logf.Log.WithName("controller_eventlogger")
	defaultFileMode  int32 = 420
	gracePeriod      int64
	eventLoggerImage = "quay.io/bakito/k8s-event-logger"
)

// Add creates a new EventLogger Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {

	podImage, ok := os.LookupEnv(c.EnvEventLoggerImage)
	if ok {
		eventLoggerImage = podImage
	}

	return &ReconcileEventLogger{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("eventlogger-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource EventLogger
	err = c.Watch(&source.Kind{Type: &eventloggerv1.EventLogger{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pod and requeue the owner EventLogger
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &eventloggerv1.EventLogger{},
	}, &deletedPredicate{})
	if err != nil {
		return err
	}
	// Watch for changes to secondary resource Secret and requeue the owner EventLogger
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &eventloggerv1.EventLogger{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileEventLogger implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileEventLogger{}

// ReconcileEventLogger reconciles a EventLogger object
type ReconcileEventLogger struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a EventLogger object and makes changes based on the state read
// and what is in the EventLogger.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileEventLogger) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	// Fetch the EventLogger cr
	cr := &eventloggerv1.EventLogger{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cr)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return r.updateCR(cr, reqLogger, err)
	}
	sec, err := secretForCR(cr)
	// Check if this Secret already exists
	secChanged, err := r.createOrReplace(cr, sec, reqLogger, func(curr runtime.Object, next runtime.Object) updateReplace {
		o1 := curr.(*corev1.Secret)
		o2 := next.(*corev1.Secret)
		if reflect.DeepEqual(o1.Data, o2.Data) {
			return no
		}
		return update
	})
	if err != nil {
		return r.updateCR(cr, reqLogger, err)
	}

	sacc, role, rb := rbacForCR(cr)
	if err != nil {
		return r.updateCR(cr, reqLogger, err)
	}
	saccChanged, err := r.createOrReplace(cr, sacc, reqLogger, nil)
	if err != nil {
		return r.updateCR(cr, reqLogger, err)
	}
	roleChanged, err := r.createOrReplace(cr, role, reqLogger, func(curr runtime.Object, next runtime.Object) updateReplace {
		o1 := curr.(*rbacv1.Role)
		o2 := next.(*rbacv1.Role)
		if reflect.DeepEqual(o1.Rules, o2.Rules) {
			return no
		}
		return update
	})
	if err != nil {
		return r.updateCR(cr, reqLogger, err)
	}
	rbChanged, err := r.createOrReplace(cr, rb, reqLogger, nil)
	if err != nil {
		return r.updateCR(cr, reqLogger, err)
	}

	// Define a new Pod object
	pod := podForCR(cr)
	// Check if this Pod already exists
	podChanged, err := r.createOrReplacePod(cr, pod, reqLogger, secChanged)
	if err != nil {
		return r.updateCR(cr, reqLogger, err)
	}

	if secChanged || saccChanged || roleChanged || rbChanged || podChanged {
		return r.updateCR(cr, reqLogger, nil)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileEventLogger) createOrReplacePod(cr *eventloggerv1.EventLogger, pod *corev1.Pod,
	reqLogger logr.Logger, replace bool) (bool, error) {

	podList := &corev1.PodList{}
	opts := []client.ListOption{
		client.InNamespace(cr.Namespace),
		client.MatchingLabels(map[string]string{
			"app":        cr.Name,
			"created-by": "eventlogger"}),
	}
	err := r.client.List(context.TODO(), podList, opts...)

	if err != nil {
		return false, err
	}

	if replace || len(podList.Items) > 1 {

		for _, p := range podList.Items {
			reqLogger.Info(fmt.Sprintf("Deleting %s", pod.Kind), "Namespace", pod.GetNamespace(), "Name", pod.GetName())
			err = r.client.Delete(context.TODO(), &p, &client.DeleteOptions{GracePeriodSeconds: &gracePeriod})
			if err != nil {
				return false, err
			}
		}
		podList = &corev1.PodList{}
	}

	if len(podList.Items) == 0 {
		// Set EventLogger cr as the owner and controller
		if err := controllerutil.SetControllerReference(cr, pod, r.scheme); err != nil {
			return false, err
		}
		reqLogger.Info(fmt.Sprintf("Creating a new %s", pod.Kind), "Namespace", pod.GetNamespace(), "Name", pod.GetName())
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func (r *ReconcileEventLogger) createOrReplace(cr *eventloggerv1.EventLogger,
	res runtime.Object,
	reqLogger logr.Logger,
	updateCheck func(curr runtime.Object, next runtime.Object) updateReplace) (bool, error) {
	query := res.DeepCopyObject()
	mo := res.(metav1.Object)
	// Check if this Resource already exists
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: mo.GetName(), Namespace: mo.GetNamespace()}, query)
	if err != nil && errors.IsNotFound(err) {
		// Set EventLogger cr as the owner and controller
		if err := controllerutil.SetControllerReference(cr, mo, r.scheme); err != nil {
			return false, err
		}

		reqLogger.Info(fmt.Sprintf("Creating a new %s", query.GetObjectKind().GroupVersionKind().Kind), "Namespace", mo.GetNamespace(), "Name", mo.GetName())
		err = r.client.Create(context.TODO(), res.(runtime.Object))
		if err != nil {
			return false, err
		}
		return true, nil

	} else if err != nil {
		return false, err
	}

	if updateCheck != nil {
		check := updateCheck(query, res)
		if check == update {
			reqLogger.Info(fmt.Sprintf("Updating %s", query.GetObjectKind().GroupVersionKind().Kind), "Namespace", mo.GetNamespace(), "Name", mo.GetName())
			err = r.client.Update(context.TODO(), res.(runtime.Object))

			if err != nil {
				return false, err
			}
			return true, nil
		} else if check == replace {
			reqLogger.Info(fmt.Sprintf("Replacing %s", query.GetObjectKind().GroupVersionKind().Kind), "Namespace", mo.GetNamespace(), "Name", mo.GetName())

			err = r.client.Delete(context.TODO(), query.(runtime.Object))

			if err != nil {
				return false, err
			}
			err = r.client.Create(context.TODO(), query.(runtime.Object))

			if err != nil {
				return false, err
			}
			return true, nil
		}
	}
	// Resource already exists
	return false, nil
}

func (r *ReconcileEventLogger) updateCR(cr *eventloggerv1.EventLogger, logger logr.Logger, err error) (reconcile.Result, error) {
	if err != nil {
		logger.Error(err, "")
		cr.Status.Error = err.Error()
	} else {
		cr.Status.Error = ""
	}
	cr.Status.LastProcessed = time.Now().Format(time.RFC3339)
	updErr := r.client.Status().Update(context.TODO(), cr)
	return reconcile.Result{}, updErr
}

// podForCR returns a pod with the same name/namespace as the cr
func podForCR(cr *eventloggerv1.EventLogger) *corev1.Pod {
	labels := map[string]string{
		"app":        cr.Name,
		"created-by": "eventlogger",
	}

	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-" + randString(),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "event-logger",
					Image:           eventLoggerImage,
					ImagePullPolicy: corev1.PullAlways,
					Env: []corev1.EnvVar{
						{
							Name: "WATCH_NAMESPACE",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.namespace",
								},
							},
						},
						{
							Name: "POD_NAME",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.name",
								},
							},
						},
						{
							Name:  c.EnvConfigFilePath,
							Value: elAbsConfigFilePath,
						},
						{
							Name:  "DEBUG_CONFIG",
							Value: "true",
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "config",
							MountPath: elAbsConfigDirPath,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							DefaultMode: &defaultFileMode,
							SecretName:  cr.Name + "-event-logger",
							Items: []corev1.KeyToPath{
								{
									Key:  elConfigFileName,
									Path: elConfigFileName,
								},
							},
						},
					},
				},
			},
			ServiceAccountName: cr.Name,
		},
	}
}

func secretForCR(cr *eventloggerv1.EventLogger) (*corev1.Secret, error) {

	conf, err := yaml.Marshal(cr.Spec)
	if err != nil {
		return nil, err
	}
	sec := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind: "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-event-logger",
			Namespace: cr.Namespace,
			Labels: map[string]string{
				"app": cr.Name,
			},
		},
		Type: "github.com/bakito/k8s-event-logger-operator",
		Data: map[string][]byte{
			elConfigFileName: conf,
		},
	}
	return sec, nil
}

func rbacForCR(cr *eventloggerv1.EventLogger) (*corev1.ServiceAccount, *rbacv1.Role, *rbacv1.RoleBinding) {
	sacc := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind: "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels: map[string]string{
				"app": cr.Name,
			},
		},
	}

	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind: "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels: map[string]string{
				"app": cr.Name,
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"events", "pods"},
				Verbs:     []string{"watch", "get", "list"},
			},
		},
	}

	rb := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind: "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels: map[string]string{
				"app": cr.Name,
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      cr.Name,
				Namespace: cr.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     cr.Name,
		},
	}

	return sacc, role, rb
}

type updateReplace string

const (
	update  updateReplace = "update"
	replace updateReplace = "replace"
	no      updateReplace = "no"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func randString() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

type deletedPredicate struct {
	predicate.Funcs
}

// Create implements Predicate
func (p deletedPredicate) Create(e event.CreateEvent) bool {
	return false
}

// Delete implements Predicate
func (p deletedPredicate) Delete(e event.DeleteEvent) bool {
	return true
}

// Update implements Predicate
func (p deletedPredicate) Update(e event.UpdateEvent) bool {
	return false
}

// Generic implements Predicate
func (p deletedPredicate) Generic(e event.GenericEvent) bool {
	return false
}