package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func runCreateTest(c client, gwAddress string) {
	testPrefix := time.Now().Unix()

	for i := range 1000 {
		name := fmt.Sprintf("%d-%d", testPrefix, i)

		defer cleanup(c, name)
		runSingleCreateTest(c, name, gwAddress)
	}
}

func runSingleCreateTest(c client, name, gwAddress string) {

	createWorkload(c, name)

	startTime, route := createRoute(c, name, "service-"+name)

	endTime := waitForReady(c, gwAddress, route, "")

	if !endTime.IsZero() {
		fmt.Printf("run %q - time to ready: %v\n", name, endTime.Sub(startTime))
	} else {
		fmt.Printf("run %q didn't become ready after: %v\n", name, time.Since(startTime))
	}
}

func waitForReady(c client, address string, route *gwv1.HTTPRoute, body string) time.Time {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 60 * time.Second,
	}

	tp := http.DefaultTransport.(*http.Transport)
	tp.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, address+":80")
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tp,
	}

	log.Print("Polling ")
	err := wait.PollUntilContextTimeout(c.ctx, time.Millisecond, 2*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://%s", string(route.Spec.Hostnames[0])), nil)

		resp, err := client.Do(req)
		if err != nil {
			return true, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Print(req.URL.String(), " not ready - status: ", resp.StatusCode)
			return false, nil
		}

		if body == "" {
			return true, nil
		}

		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Panicf("error reading body: %v", err)
		}

		if !strings.Contains(string(bytes), body) {
			log.Print(req.URL.String(), " body not ready")
			return false, nil
		}
		return true, nil
	})

	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		log.Print(err)
		return time.Time{}
	} else if err != nil {
		log.Panicf("failed to query readiness: %s", err)
	}

	return time.Now()
}
