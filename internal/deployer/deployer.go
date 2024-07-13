package deployer

import (
	"context"
	"fmt"

	"sync-algo/internal/config"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Deployer struct {
	clientset *kubernetes.Clientset
	cfg       *config.Kubernates
}

func New(cfg *config.Kubernates) (*Deployer, error) {
	config, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	return &Deployer{
		clientset: clientset,
	}, nil
}

func (d *Deployer) CreatePod(name string) error {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  name,
					Image: d.cfg.ConteinerName,
				},
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := d.clientset.CoreV1().Pods("default").Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create pod: %w", err)
	}

	return nil
}

func (d *Deployer) DeletePod(name string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := d.clientset.CoreV1().Pods("default").Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}

	return nil
}

func (d *Deployer) GetPodList() ([]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pods, err := d.clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var podNames []string
	for _, pod := range pods.Items {
		podNames = append(podNames, pod.Name)
	}

	return podNames, nil
}
