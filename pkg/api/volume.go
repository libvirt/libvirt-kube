package api

import (
	"encoding/base64"
	"fmt"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kubeapiv1 "k8s.io/client-go/pkg/api/v1"
)

func getVolumeClaimVolumeName(clientset *kubernetes.Clientset, name, namespace string) (string, error) {
	options := metav1.GetOptions{}
	glog.V(1).Infof("Querying PVC %s/%s", namespace, name)
	pvc, err := clientset.CoreV1().PersistentVolumeClaims(namespace).Get(name, options)
	if err != nil {
		return "", err
	}

	return pvc.Spec.VolumeName, nil
}

func GetVolumeSpec(clientset *kubernetes.Clientset, name, namespace string) (string, *kubeapiv1.PersistentVolumeSpec, error) {
	volname, err := getVolumeClaimVolumeName(clientset, name, namespace)
	if err != nil {
		return "", nil, err
	}

	glog.V(1).Infof("Querying PV %s", volname)
	options := metav1.GetOptions{}
	pv, err := clientset.CoreV1().PersistentVolumes().Get(volname, options)
	if err != nil {
		return "", nil, err
	}

	return volname, &pv.Spec, nil
}

func GetVolumeRBDKey(clientset *kubernetes.Clientset, namespace string, src *kubeapiv1.RBDVolumeSource) ([]byte, error) {
	if src.SecretRef != nil {
		key64, err := GetSecretValue(clientset, src.SecretRef.Name, namespace, "kubernetes.io/rbd", "key")
		if err != nil {
			return []byte{}, err
		}

		key, err := base64.StdEncoding.DecodeString(string(key64))
		if err != nil {
			return []byte{}, err
		}
		return key, nil
	} else {
		return []byte{}, fmt.Errorf("No secret defined for volume")
	}
}
