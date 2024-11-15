package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	namespace string
	gateway   string
	test      string
)

func main() {
	flag.StringVar(&namespace, "namespace", "default", "default namespace to create test resource")
	flag.StringVar(&gateway, "gateway", "gateway", "default gateway routes are attached to")
	flag.StringVar(&test, "test", "create", "test to run 'create|update'")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	client := createClient()
	client.ctx = ctx

	g, err := client.GatewayV1().Gateways(namespace).Get(client.ctx, gateway, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("failed to get gateway %q: %s", gateway, err)
	}

	if len(g.Status.Addresses) < 1 {
		log.Fatalf("Gateway %q doesn't have an address", gateway)
	}

	gwAddress := g.Status.Addresses[0].Value

	switch test {
	case "create":
		log.Println("Running Create Test")
		runCreateTest(client, gwAddress)
	case "update":
		log.Println("Running Update Test")
		runUpdateTest(client, gwAddress)
	}
}
