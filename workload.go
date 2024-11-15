package main

import (
	"context"
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"

	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var (
	podTemplate = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:            "web",
				Image:           "ghcr.io/knative/helloworld-go:latest",
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{{
					Name:  "TARGET",
					Value: "",
				}},
				Ports: []corev1.ContainerPort{{
					Name:          "web",
					ContainerPort: 8080,
					Protocol:      corev1.ProtocolTCP,
				}},
			}},
		},
	}

	serviceTemplate = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{},
			Ports: []corev1.ServicePort{{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Protocol:   corev1.ProtocolTCP,
			}},
		},
	}

	port80 = gwv1.PortNumber(80)

	routeTemplate = &gwv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{},
		},
		Spec: gwv1.HTTPRouteSpec{
			CommonRouteSpec: gwv1.CommonRouteSpec{
				ParentRefs: []gwv1.ParentReference{{
					Name: "",
				}},
			},
			Hostnames: []gwv1.Hostname{},
			Rules: []gwv1.HTTPRouteRule{{
				BackendRefs: []gwv1.HTTPBackendRef{{
					BackendRef: gwv1.BackendRef{
						BackendObjectReference: gwv1.BackendObjectReference{
							Port:      &port80,
							Namespace: (*gwv1.Namespace)(&namespace),
						},
					},
				}},
			}},
		},
	}
)

func createWorkload(c client, name string) {
	pod := podTemplate.DeepCopy()
	pod.Name = name
	pod.Labels["name"] = name
	pod.Spec.Containers[0].Env[0].Value = name

	service := serviceTemplate.DeepCopy()
	service.Name = "service-" + name
	service.Spec.Selector["name"] = name

	_, err := c.CoreV1().Pods(namespace).Create(c.ctx, pod, metav1.CreateOptions{})
	if err != nil {
		log.Panicf("failed to create pod: %s", err)
	}

	_, err = c.CoreV1().Services(namespace).Create(c.ctx, service, metav1.CreateOptions{})
	if err != nil {
		log.Panicf("failed to create service: %s", err)
	}

	err = wait.PollUntilContextTimeout(c.ctx, 10*time.Millisecond, 10*time.Second, true, func(ctx context.Context) (done bool, err error) {
		e, err := c.CoreV1().Endpoints(namespace).Get(ctx, service.Name, metav1.GetOptions{})
		if err != nil && apierrors.IsNotFound(err) {
			return false, nil
		} else if err != nil {
			return true, err
		}

		for _, subset := range e.Subsets {
			if len(subset.Addresses) > 0 {
				return true, nil
			}
		}
		return false, nil
	})

	if err != nil {
		log.Panicf("service failed to observe ready endpoint: %s", err)
	}
}

func createRoute(c client, name, backend string) (time.Time, *gwv1.HTTPRoute) {
	hostname := gwv1.Hostname(name + ".example.com")
	route := routeTemplate.DeepCopy()

	route.Name = name
	route.Spec.CommonRouteSpec.ParentRefs[0].Name = gwv1.ObjectName(gateway)
	route.Spec.Hostnames = append(route.Spec.Hostnames, hostname)
	route.Spec.Rules[0].BackendRefs[0].BackendObjectReference.Name = gwv1.ObjectName(backend)

	r, err := c.GatewayV1().HTTPRoutes(namespace).Create(c.ctx, route, metav1.CreateOptions{})

	if err != nil {
		log.Panicf("failed to create route: %s", err)
	}

	return time.Now(), r
}

func updateRoute(c client, name, backend string) (time.Time, *gwv1.HTTPRoute) {
	route, err := c.GatewayV1().HTTPRoutes(namespace).Get(c.ctx, name, metav1.GetOptions{})
	if err != nil {
		log.Panicf("failed to fetch route: %s", err)
	}

	route.Spec.Rules[0].BackendRefs[0].BackendObjectReference.Name = gwv1.ObjectName(backend)

	r, err := c.GatewayV1().HTTPRoutes(namespace).Update(c.ctx, route, metav1.UpdateOptions{})

	if err != nil {
		log.Panicf("failed to create route: %s", err)
	}

	return time.Now(), r
}

func cleanup(c client, name string) {
	c.GatewayV1().HTTPRoutes(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	c.CoreV1().Pods(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	c.CoreV1().Services(namespace).Delete(context.Background(), "service-"+name, metav1.DeleteOptions{})
}
