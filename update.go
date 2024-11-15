package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func runUpdateTest(c client, gwAddress string) {
	testPrefix := time.Now().Unix()

	routeName := fmt.Sprintf("%d", testPrefix)

	defer c.GatewayV1().HTTPRoutes(namespace).Delete(context.Background(), routeName, metav1.DeleteOptions{})

	// Initial workload
	name := fmt.Sprintf("%d-%d", testPrefix, 0)
	createWorkload(c, name)

	startTime, route := createRoute(c, routeName, "service-"+name)
	endTime := waitForReady(c, gwAddress, route, name)

	if !endTime.IsZero() {
		fmt.Printf("run %q - time to ready: %v\n", name, endTime.Sub(startTime))
	} else {
		fmt.Printf("run %q didn't become ready after: %v\n", name, time.Since(startTime))
	}

	rate := vegeta.Rate{Freq: 50, Per: time.Second}

	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    fmt.Sprintf("http://%v.example.com/", routeName),
	})
	attacker := vegeta.NewAttacker(vegeta.Client(gwClient(gwAddress)))

	var metrics vegeta.Metrics

	go func() {
		for res := range attacker.Attack(targeter, rate, 0, "Big Bang!") {
			metrics.Add(res)
		}
		metrics.Close()

		bytes, err := json.MarshalIndent(metrics, "", " ")

		fmt.Println(string(bytes), " ", err)
	}()

	defer cleanup(c, name)

	for i := range 100 {
		name := fmt.Sprintf("%d-%d", testPrefix, i+1)

		createWorkload(c, name)
		defer cleanup(c, name)

		startTime, route := updateRoute(c, routeName, "service-"+name)
		endTime := waitForReady(c, gwAddress, route, name)

		if !endTime.IsZero() {
			fmt.Printf("run %q - time to update: %v\n", name, endTime.Sub(startTime))
		} else {
			fmt.Printf("run %q didn't update after: %v\n", name, time.Since(startTime))
		}

	}

	attacker.Stop()
}

func gwClient(address string) *http.Client {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 60 * time.Second,
	}

	tp := http.DefaultTransport.(*http.Transport)
	tp.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, address+":80")
	}

	return &http.Client{
		Timeout:   10 * time.Second,
		Transport: tp,
	}
}
