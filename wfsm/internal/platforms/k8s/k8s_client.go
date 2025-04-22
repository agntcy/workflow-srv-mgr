package k8s

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func getK8sClient() (*kubernetes.Clientset, error) {
	factory := cmdutil.NewFactory(genericclioptions.NewConfigFlags(true))
	return factory.KubernetesClientSet()
}
