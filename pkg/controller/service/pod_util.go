package service

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (r *ReconcileService) podIPs(pods []corev1.Pod) ([]string, error) {
	ips := map[string]bool{}

	for _, pod := range pods {
		if pod.Spec.NodeName == "" || pod.Status.PodIP == "" {
			continue
		}
		if !isPodReady(&pod) {
			continue
		}

		node, err := r.getNode(pod.Spec.NodeName)
		if errors.IsNotFound(err) {
			continue
		} else if err != nil {
			return nil, err
		}

		var internal string
		var external string
		for _, addr := range node.Status.Addresses {
			// Prefer external address, if not set on node use internal address
			if addr.Type == corev1.NodeExternalIP {
				external = addr.Address
			}

			if addr.Type == corev1.NodeInternalIP {
				internal = addr.Address
			}
		}
		if external != "" {
			ips[external] = true
		} else {
			ips[internal] = true
		}
	}

	var ipList []string
	for k := range ips {
		ipList = append(ipList, k)
	}
	return ipList, nil
}

func isPodReady(pod *corev1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == "Ready" && c.Status == "True" {
			return true
		}
	}

	return false
}
