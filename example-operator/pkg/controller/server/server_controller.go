package server

import (
	"context"
	"strconv"
	"math/rand"

	tonyv1alpha1 "github.com/pod-operator/tony-operator/pkg/apis/tony/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
//	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
//	"k8s.io/client-go/tools/cache"
//	"k8s.io/apimachinery/pkg/watch"
//	"k8s.io/client-go/util/workqueue"
//	"sigs.k8s.io/controller-runtime/pkg/client/config"
//	"k8s.io/client-go/kubernetes"
//	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var log = logf.Log.WithName("controller_server")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Server Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileServer{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("server-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Server
	err = c.Watch(&source.Kind{Type: &tonyv1alpha1.Server{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Server
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tonyv1alpha1.Server{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileServer implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileServer{}

// ReconcileServer reconciles a Server object
type ReconcileServer struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Server object and makes changes based on the state read
// and what is in the Server.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileServer) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Server")
	// Fetch the Server instance
	instance := &tonyv1alpha1.Server{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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
	// Check if this Pod already exists.
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	//get the desired number of pod.
	count := instance.Spec.Count
	//try to match the number defined in specification.
        //List the pod.
	podList := &corev1.PodList{}
	//get the label selector.
	labelSelector := labels.SelectorFromSet(labelsForCustomerResource(instance.Name))
	listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
	 //get the current pod names.
        r.client.List(context.TODO(), listOps, podList)
        currentNumber := len(getPodsNames(podList.Items))
	//Print the Information:
	log.Info("Current Number is: ", strconv.Itoa(currentNumber), strconv.Itoa(count))

	if count > currentNumber {
		//Number the pods.
	        i:= rand.Intn(10000)
		//create a new pod.
		pod := NewPodForCR(instance, i)
		// Set Server instance as the owner and controller, enable to watch the pods.
		if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
			reqLogger.Error(err, "unable to reference on new pod.")
			return reconcile.Result{}, err
		}
		err = r.client.Create(context.TODO(), pod)
		// print out the information.
		if err != nil {
	                return reconcile.Result{}, err
		} else {
		        reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		}
	    // Pod created successfully - don't requeue
	} else if count < currentNumber {
            log.Info("******Delete begins with here.*******")
	    log.Info("Current Number is: ", strconv.Itoa(currentNumber), strconv.Itoa(count))
	    //delete pods begins here
	    for i := currentNumber; i > count; i-- {
	        //generate a simluation pod.
                simulationPod := podList.Items[i - 1].DeepCopyObject()
		err = r.client.Delete(context.TODO(), simulationPod)
		//print info
		r.client.List(context.TODO(), listOps, podList)
                currentNumber = len(getPodsNames(podList.Items))
                //Print the Information:
                log.Info("After deleting, current Number is: ", strconv.Itoa(currentNumber), strconv.Itoa(count))

	        if err != nil {
		     return reconcile.Result{}, err
	        } else {
		        log.Info("A pod is deleted successfully.")
			}
	    }
        } else {
	   log.Info ("Nothing to do....")
	}
  return reconcile.Result{Requeue: true}, nil
}

/**
 * The main purpose of this function is to define and create the pods via Go.
 * @param cr the parameter defined in the customer resource.
 * @param number the number used to set the pods apart.
 * @return pod a new pod.
 */
func NewPodForCR(cr *tonyv1alpha1.Server, number int) *corev1.Pod {
	labels := labelsForCustomerResource(cr.Name)
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod-" + strconv.Itoa(number),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
				      Name:    cr.Name,
			              Image:   cr.Spec.Image,
				      Ports: []corev1.ContainerPort {
					{//Array begins here.
					    ContainerPort: cr.Spec.Port,
					    Name: cr.Name,
					},
				      },
				},
			},
		},
	}
}

/**
 * This function is to return the all the names of the pods from an array.
 * @param  pods PodList struct in json format with `json:"items"`.
 * @return podNames an array of all the pods names.
 */
func getPodsNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
	    //Add the pod name to the list
	    podNames = append(podNames, pod.Name)
	}
	return podNames
}

func labelsForCustomerResource(name string) map[string]string {
	return map[string]string{"app": "Example-Operator", "exampleoperator_cr": name}
}
