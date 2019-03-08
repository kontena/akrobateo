package service

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceWrangler struct {
	client  client.Client
	service corev1.Service
}

func (sw *ServiceWrangler) ExistingIPs() []string {
	var ips []string

	for _, ingress := range sw.service.Status.LoadBalancer.Ingress {
		if ingress.IP != "" {
			ips = append(ips, ingress.IP)
		}
	}

	return ips
}

func (sw *ServiceWrangler) FindPods() (*corev1.PodList, error) {
	selector, err := labels.Parse(svcNameLabel + "=" + sw.service.Name)
	if err != nil {
		return nil, err
	}
	opts := &client.ListOptions{
		Namespace:     sw.service.Namespace,
		LabelSelector: selector,
	}
	pods := &corev1.PodList{}
	err = sw.client.List(context.TODO(), opts, pods)

	return pods, err
}

func (sw *ServiceWrangler) UpdateAddresses(ips []string) error {
	svc := sw.service.DeepCopy()
	svc.Status.LoadBalancer.Ingress = nil
	for _, ip := range ips {
		svc.Status.LoadBalancer.Ingress = append(svc.Status.LoadBalancer.Ingress, corev1.LoadBalancerIngress{
			IP: ip,
		})
	}

	return sw.client.Status().Update(context.TODO(), svc)
}
