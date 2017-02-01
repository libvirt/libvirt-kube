package api

import (
	"fmt"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kubeapiv1 "k8s.io/client-go/pkg/api/v1"
)

func GetSecretValue(clientset *kubernetes.Clientset, name, namespace, stype, field string) ([]byte, error) {
	glog.V(1).Infof("Querying secret %s/%s", namespace, name)
	options := metav1.GetOptions{}
	sec, err := clientset.CoreV1().Secrets(namespace).Get(name, options)
	if err != nil {
		return []byte{}, err
	}

	if sec.Type != kubeapiv1.SecretType(stype) {
		return []byte{}, fmt.Errorf("Secret %s/%s type is %s but want %s",
			namespace, name, sec.Type, stype)
	}

	val, ok := sec.Data[field]
	if !ok {
		return []byte{}, fmt.Errorf("Secret %s/%s missing field %s", namespace, name, field)
	}

	return val, nil
}
