package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/thomastaylor312/advanced-helm-demos/controller/pkg/generated/clientset/versioned"
	informers "github.com/thomastaylor312/advanced-helm-demos/controller/pkg/generated/informers/externalversions"
)

func main() {
	ctx := setupSignalHandler()
	cfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err)
	}

	helmClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building example clientset: %s", err)
	}

	helmInformerFactory := informers.NewSharedInformerFactory(helmClient, time.Second*30)

	controller, err := NewController(helmClient, helmInformerFactory.Helmcontroller().V1alpha1().Helms())
	if err != nil {
		log.Fatalf("Unable to construct controller: %s", err)
	}

	helmInformerFactory.Start(ctx.Done())

	if err = controller.Run(1, ctx.Done()); err != nil {
		log.Fatalf("Error running controller: %s", err.Error())
	}
}

func setupSignalHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cancel()
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return ctx
}
