package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/resource"
)

type clusterReleaseStatus struct {
	KindCount     map[string]uint
	TotalReleases uint
	NumReady      uint
	NumNotReady   uint
}

func main() {
	settings := cli.New()

	actionConfig := new(action.Configuration)
	// Passing an empty namespace means we want to list across all namespaces
	if err := actionConfig.Init(settings.RESTClientGetter(), "", os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Fatalf("unable to initialize Helm config: %s", err)
	}

	var clusterStatus clusterReleaseStatus

	client := action.NewList(actionConfig)
	// Make sure we are only listing deployed releases
	client.Deployed = true

	releases, err := client.Run()
	if err != nil {
		log.Fatalf("error while trying to list releases: %s", err)
	}

	clusterStatus.TotalReleases = uint(len(releases))

	var allTheObjects kube.ResourceList
	for _, rel := range releases {
		// We can't just use the client available on actionConfig because it is
		// using the default namespace, so construct a new one here
		namespacedClient := kube.New(kube.GetConfig(settings.KubeConfig, settings.KubeContext, rel.Namespace))
		objs, err := namespacedClient.Build(bytes.NewBufferString(rel.Manifest), false)
		if err != nil {
			log.Fatalf("unable to build kubernetes objects from release manifest: %s", err)
		}
		allTheObjects = append(allTheObjects, objs...)
	}

	var wg sync.WaitGroup
	wg.Add(len(allTheObjects))

	// Use a normal for loop to avoid pointer issues with a range loop
	// Here we are doing all the calls to the k8s API concurrently
	for i := 0; i < len(allTheObjects); i++ {
		go func(info *resource.Info) {
			defer wg.Done()
			err := info.Get()
			if err != nil {
				// This is a demo, so no error handling here
				log.Printf("[ERROR] unable to fetch object data from Kubernetes: %s", err)
				return
			}
		}(allTheObjects[i])
	}
	wg.Wait()

	// Init the map beforehand
	clusterStatus.KindCount = make(map[string]uint)
	for _, info := range allTheObjects {
		// First increment the kind count
		gvk := info.Object.GetObjectKind().GroupVersionKind()
		clusterStatus.KindCount[fmt.Sprintf("%s/%s/%s", gvk.GroupVersion().Group, gvk.GroupVersion().Version, gvk.GroupKind().Kind)]++

		// Now match on a few object kinds to get statuses. This is super naive,
		// so see the wait logic in the Helm codebase for better examples of
		// checking for ready state
		var isSupported, isReady bool
		switch value := kube.AsVersioned(info).(type) {
		case *corev1.PersistentVolumeClaim:
			isSupported = true
			isReady = value.Status.Phase == corev1.ClaimBound
		case *appsv1.Deployment:
			isSupported = true
			var expectedReplicas int32
			if value.Spec.Replicas == nil {
				expectedReplicas = 1
			} else {
				expectedReplicas = *value.Spec.Replicas
			}
			isReady = value.Status.ReadyReplicas == expectedReplicas
		case *appsv1.StatefulSet:
			isSupported = true
			var expectedReplicas int32
			if value.Spec.Replicas == nil {
				expectedReplicas = 1
			} else {
				expectedReplicas = *value.Spec.Replicas
			}
			isReady = value.Status.ReadyReplicas == expectedReplicas
		}

		if isSupported && isReady {
			clusterStatus.NumReady++
		} else if isSupported && !isReady {
			clusterStatus.NumNotReady++
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	enc.Encode(clusterStatus)
}
