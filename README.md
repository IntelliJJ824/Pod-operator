# Pod-operator
# Project status: beta
Major planned features have been completed. However, potential extension and functions have not been confirmed. 
This project is aim to understand the operator as beginner level. In the following sections, there would be a guide to 
write this operator steps by steps.

# Overview 
The pod operator manages the pods applied to Kubernetes and automates tasks related to operating the pods.
  - Create and Delete 
  - Resize
  - Fail over

# DEMO
 ### Create and Destroy Pods
 Three pods would be created as requirement specified in the tony_v1alpha1_server_cr.yaml.
 ```sh
    $ kubectl apply -f /root/go/src/github.com/pod-operator/tony-operator/deploy/crds/tony_v1alpha1_server_cr.yaml 
    $ kubectl get pods
    NAME                           READY    STATUS    RESTARTS   AGE
    example-server-pod-1137         1/1     Running   0          9m22s
    example-server-pod-3033         1/1     Running   0          9m22s
    example-server-pod-3133         1/1     Running   0          9m22s
    tony-operator-59c79dbd6-dpj5q   1/1     Running   0          4h25m
 ```
 ### Resize the pods
 In tony_v1alpha1_server_cr.yaml the initial cluster size is 3. Modify the file and change <b>count</b> from 3 to 1.
 ```sh
    $ cat /root/go/src/github.com/pod-operator/tony-operator/deploy/crds/tony_v1alpha1_server_cr.yaml
    apiVersion: "tony.stark.com/v1alpha1"
    kind: "Server"
    metadata:
      name: "example-server"
      namespace: default
    spec:
      # Add fields here
      count: 1
      group: "example-app"
      image: docker.io/zz35/tony-server
      port: 80
 ```
 ```sh
    $ kubectl apply -f /root/go/src/github.com/pod-operator/tony-operator/deploy/crds/tony_v1alpha1_server_cr.yaml
    $ kubectl get pods
    NAME                            READY   STATUS    RESTARTS   AGE
    example-server-pod-1137         1/1     Running   0          30m
    tony-operator-59c79dbd6-dpj5q   1/1     Running   0          4h46m
```
### Failover the Pods
This function is used to create a new pod for the cluster when a running pod is deleted by kubectl.
```sh 
   $ kubectl delete pod example-server-pod-1137
   pod "example-server-pod-1137" deleted
   $ kubectl get pods
   NAME                            READY   STATUS    RESTARTS   AGE
   example-server-pod-7425         1/1     Running   0          26s
   tony-operator-59c79dbd6-dpj5q   1/1     Running   0          5h
```

# Prerequisites
  - Kubernetes 1.11.3+.
  - dep version v0.5.0+.
  - go version v1.12+.
  - git
  - operator-SDK v0.8.1
  - ubuntu v18+ (TBC)

# Quick Start  
## Installation of Operator SDK
  Follow the steps in the [installation guide](https://github.com/operator-framework/operator-sdk/tree/v0.8.1) 
  to learn how to install the Operator SDK.  
  Please <b>carefully read</b> the following things:
  - Check the $GOPATH and $GOROOT
    (Mine: GOPATH="/root/go")
    
  - It would be the best to install Operator SDK under $GOPATH/src directory. 
    <br>( i.e ~/go/src/github.com/operator-framework )
    
  - One step, make dep, mentioned in the installation guide is not compulsory
  - It is common to receive error/warning messages during the installation. 
  - As long as the installation exits as normal, ignore those errors/warnings during the process.
  - Ubuntu under v18.0.0 seems not to work for Operator SDK.

## Create and deploy a pod-operator
### 1. Describe the Customer Resource
Modify the API: ~/go/src/github.com/pod-operator/tony-operator/pkg/apis/tony/v1alpha1/server_types.go<br>
It is used to get some custom field for our CR yaml file. Another things is that every time server_types.go is revised, 
please do not forget to type <b>operator-sdk generate k8s</b> in terminal to create the deep-copy codes. 
```GO
type ServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Count int `json:"count"`
	Group string `json:"group"`
	Image string `json:"image"`
	Port int32 `json:"port"`
}
``` 
```GO
type ServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	PodNames []string `json:"podnames"`
	AppGroup string `json:"appgroup"`
}
```
### 2. Write the Logic to the Operator
In this part, ~/go/src/github.com/pod-operator/tony-operator/pkg/controller/server/server_controller.go is modified. 
The machinism of the operator is designed in here. 

##### 2.1 Define a function to create a pod
The main purpose of this function is to define and create the pods via Go.
```GO
/**
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
```
##### 2.2 Retrieve all the pods
This function is to retrieve all the pods related to the customer resource.
```GO
/**
 * @param pods the list of pods.
 * @return podNames an array with all the pod names related to the custom resource.
 */
func getPodsNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
	    //Add the pod name to the list
	    podNames = append(podNames, pod.Name)
	}
	return podNames
}
```



##### 2.3 Logic of reconciliation 
This is the most essential part of the operator. The logic of the controller is defined in the following function. 
However, one things that need to be investigated is the multi-thread problem.
```GO
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
	//try to match the number defined in spce.
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
		//controllerutil.SetControllerReference(instance, pod, r.scheme)
		err = r.client.Create(context.TODO(), pod)
                // print out the information.
		if err != nil {
	                return reconcile.Result{}, err
		} else {
		        reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		}
	    // Pod created successfully - don't requeue
	    // return reconcile.Result{}, nil
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
	    //return reconcile.Result{}, nil
        } else {
	   log.Info ("Nothing to do....")
	}
  return reconcile.Result{Requeue: true}, nil
}
```
### 3. Deploy the operator
This is the last step to complete the process of operator deployment. 
You could follow the instructions mentioned in the Operator SDK installation page or doing the following things.
```Bash
    $ pwd
    /root/go/src/github.com/pod-operator/tony-operator
    # Delete all the deployments and pods.
    $ sh ./bashScript/deleteDeployment.sh
    # Build the operator.
    $ sh ./bashScript/buildOperator.sh
    # Create related deployments including CRD.
    $ sh ./bashScript/createDeployment.sh
    # Deploy the operator.
    $ sh ./bashScript/createOperator.sh
    # Write and set the specification for the desired status (CR).
    $ sh ./bashScript/CreateCR.sh
``` 


