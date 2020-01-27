package main

import (
	"fmt"
	"log"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	helmv1alpha1 "github.com/thomastaylor312/advanced-helm-demos/controller/pkg/apis/helmcontroller/v1alpha1"
	clientset "github.com/thomastaylor312/advanced-helm-demos/controller/pkg/generated/clientset/versioned"
	informers "github.com/thomastaylor312/advanced-helm-demos/controller/pkg/generated/informers/externalversions/helmcontroller/v1alpha1"
	listers "github.com/thomastaylor312/advanced-helm-demos/controller/pkg/generated/listers/helmcontroller/v1alpha1"
)

const (
	helmStorageDriver = "secrets"
)

// Controller is the controller implementation for Helm resources. This is
// adapted from the Kubernetes Sample Controller found here:
// https://github.com/kubernetes/sample-controller
type Controller struct {
	// clientset is a clientset for our own API group
	clientset clientset.Interface

	helmLister listers.HelmLister
	helmSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	// The base chart used in creating new releases. This should only ever be
	// read and not written to.
	baseChart *chart.Chart
}

func NewController(helmclientset clientset.Interface, helmInformer informers.HelmInformer) (*Controller, error) {
	// Load the base chart from disk
	chrt, err := loader.Load("/charts/base")
	if err != nil {
		return nil, err
	}

	controller := &Controller{
		clientset:  helmclientset,
		helmLister: helmInformer.Lister(),
		helmSynced: helmInformer.Informer().HasSynced,
		workqueue:  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Helm"),
		baseChart:  chrt,
	}

	log.Printf("Setting up event handlers")
	// Set up an event handler for when Foo resources change
	helmInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueHelm,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueHelm(new)
		},
		DeleteFunc: controller.deleteRelease,
	})

	return controller, nil
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	log.Printf("Starting Helm controller")

	// Wait for the caches to be synced before starting workers
	log.Printf("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.helmSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Printf("Starting workers")
	// Launch two workers to process resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	log.Printf("Started workers")
	<-stopCh
	log.Printf("Shutting down workers")

	return nil
}

func (c *Controller) deleteRelease(obj interface{}) {
	typedObj, ok := obj.(*helmv1alpha1.Helm)
	if !ok {
		log.Printf("[WARN] Received object of type %T instead of a Helm object", obj)
		return
	}

	configOptions := genericclioptions.NewConfigFlags(false)
	configOptions.Namespace = &typedObj.Namespace
	actionConfig := new(action.Configuration)
	// Passing an empty namespace means we want to list across all namespaces
	if err := actionConfig.Init(configOptions, typedObj.Namespace, helmStorageDriver, log.Printf); err != nil {
		log.Printf("[ERROR] Unable to initialize Helm config for %s/%s: %s", typedObj.Namespace, typedObj.Name, err)
		return
	}

	log.Printf("Starting delete of release %s in namespace %s", typedObj.Name, typedObj.Namespace)
	client := action.NewUninstall(actionConfig)
	if _, err := client.Run(typedObj.Name); err != nil {
		log.Printf("[ERROR] Unable to delete Helm release: %s", err)
	}
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		log.Printf("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Helm resource with this namespace/name
	helm, err := c.helmLister.Helms(namespace).Get(name)
	if err != nil {
		// The Helm resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("helm '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	if helm.Spec.ImageName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: image name must be specified", key))
		return nil
	}

	// Create the Helm action config
	configOptions := genericclioptions.NewConfigFlags(false)
	configOptions.Namespace = &helm.Namespace
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(configOptions, helm.Namespace, helmStorageDriver, log.Printf); err != nil {
		log.Printf("[ERROR] Unable to initialize Helm config for %s/%s: %s", helm.Namespace, helm.Name, err)
		return err
	}

	// Get the release associated with the given Helm resource
	var rel *release.Release
	chartCopy := *c.baseChart
	var numReplicas int32 = 1
	if helm.Spec.Replicas != nil {
		numReplicas = *helm.Spec.Replicas
	}
	rel, err = action.NewGet(actionConfig).Run(helm.Name)
	// If the resource doesn't exist, we'll create it
	if err == driver.ErrReleaseNotFound {
		installClient := action.NewInstall(actionConfig)
		installClient.ReleaseName = helm.Name

		rel, err = installClient.Run(&chartCopy, map[string]interface{}{
			"replicaCount": numReplicas,
			"image": map[string]interface{}{
				"name": helm.Spec.ImageName,
			},
		})
	} else if err != nil {
		// An error during get means we can't assume anything about the status
		// of the release, so just error out for now
		return err
	} else {
		// Upgrade the release if something has changed
		vals, valErr := chartutil.CoalesceValues(rel.Chart, rel.Config)
		if valErr != nil {
			return valErr
		}
		// Why yes, this is an ugly type assertion. What else are demos for?
		if !(vals["image"].(map[string]interface{})["name"].(string) == helm.Spec.ImageName && numReplicas == int32(vals["replicaCount"].(float64))) {
			rel, err = action.NewUpgrade(actionConfig).Run(helm.Name, &chartCopy, map[string]interface{}{
				"replicaCount": numReplicas,
				"image": map[string]interface{}{
					"name": helm.Spec.ImageName,
				},
			})
		}
	}

	// Update the status block of the Helm resource to reflect the
	// current state of the world
	statusErr := c.updateHelmStatus(helm, rel, err)
	if statusErr != nil {
		return statusErr
	}

	// If an error occurred during install or update, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) updateHelmStatus(helm *helmv1alpha1.Helm, rel *release.Release, e error) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	helmCopy := helm.DeepCopy()
	var relStatus release.Status
	if rel == nil {
		relStatus = release.StatusUnknown
	} else {
		relStatus = rel.Info.Status
	}
	helmCopy.Status.ReleaseStatus = relStatus
	message := "Release deployed"
	if e != nil {
		message = e.Error()
	}
	helmCopy.Status.Message = message
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.clientset.HelmcontrollerV1alpha1().Helms(helm.Namespace).Update(helmCopy)
	return err
}

// enqueueHelm takes a Helm resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Helm.
func (c *Controller) enqueueHelm(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}
