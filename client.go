package main

import (
	"context"
	"log"

	kclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	gclient "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
)

type (
	kClient = *kclient.Clientset
	gClient = *gclient.Clientset

	client struct {
		kClient
		gClient

		ctx context.Context
	}
)

func createClient() client {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		overrides,
	).ClientConfig()

	if err != nil {
		log.Fatalf("failed to create client config: %s", err)
	}

	config.QPS = 100
	config.Burst = 100

	kset, err := kclient.NewForConfig(config)

	if err != nil {
		log.Fatalf("failed to create client: %s", err)
	}

	gset, err := gclient.NewForConfig(config)

	if err != nil {
		log.Fatalf("failed to create client: %s", err)
	}

	return client{
		kClient: kset,
		gClient: gset,
	}
}
