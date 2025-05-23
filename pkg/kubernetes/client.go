package kubernetes

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset *kubernetes.Clientset
}

func NewClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load Kubernetes configuration: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	return &Client{clientset: clientset}, nil
}

func (c *Client) GetNamespaces() ([]string, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	var namespaceNames []string
	for _, ns := range namespaces.Items {
		namespaceNames = append(namespaceNames, ns.Name)
	}
	return namespaceNames, nil
}

func (c *Client) GetPods(namespace string) ([]string, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	var podNames []string
	for _, pod := range pods.Items {
		podNames = append(podNames, pod.Name)
	}
	return podNames, nil
}

func (c *Client) GetServices(namespace string) ([]string, error) {
	services, err := c.clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %v", err)
	}

	var serviceNames []string
	for _, service := range services.Items {
		serviceNames = append(serviceNames, service.Name)
	}
	return serviceNames, nil
}

func (c *Client) GetPodByIP(ip string) (string, string, error) {
    pods, err := c.clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
        FieldSelector: fmt.Sprintf("status.podIP=%s", ip),
    })
    if err != nil {
        return "", "", err
    }
    if len(pods.Items) > 0 {
        return pods.Items[0].Name, pods.Items[0].Namespace, nil
    }
    return "", "", fmt.Errorf("no pod found with IP %s", ip)
}

func (c *Client) GetServiceByIP(ip string) (string, string, error) {
    services, err := c.clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{
        FieldSelector: fmt.Sprintf("spec.clusterIP=%s", ip),
    })
    if err != nil {
        return "", "", err
    }
    if len(services.Items) > 0 {
        return services.Items[0].Name, services.Items[0].Namespace, nil
    }
    return "", "", fmt.Errorf("no service found with IP %s", ip)
}