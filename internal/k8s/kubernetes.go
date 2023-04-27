package k8s

import (
	"context"

	"github.com/TwiN/go-color"
	"github.com/mxcd/gitops-cli/internal/secret"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateSecret(s *secret.Secret, clusterName string) error {
	if s.Type == "ConfigMap" {
		return CreateK8sConfigMap(s, clusterName)
	} else {
		return CreateK8sSecret(s, clusterName)
	}
}

func CreateK8sConfigMap(s *secret.Secret, clusterName string) error {
	clusterClient, err := GetClient(clusterName)
	if err != nil {
		return err
	}
	clientset := clusterClient.Clientset

	log.Trace("Creating ConfigMap ", s.Name, " in namespace ", s.Namespace)
	k8sConfigMap, err := clientset.CoreV1().ConfigMaps(s.Namespace).Create(context.Background(), &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.Name,
			Namespace: s.Namespace,
			Annotations: map[string]string{
				"gitops.mxcd.de/secret-id": s.ID,
			},
		},
		Data: s.Data,
	}, metav1.CreateOptions{})
	log.Trace(k8sConfigMap)
	if err != nil {
		return err
	}
	println(s.Namespace, "/", s.Name, color.InGreen(" created"))
	return err
}

func CreateK8sSecret(s *secret.Secret, clusterName string) error {
	clusterClient, err := GetClient(clusterName)
	if err != nil {
		return err
	}
	clientset := clusterClient.Clientset

	log.Trace("Creating Secret ", s.Name, " in namespace ", s.Namespace)
	k8sSecret, err := clientset.CoreV1().Secrets(s.Namespace).Create(context.Background(), &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.Name,
			Namespace: s.Namespace,
			Annotations: map[string]string{
				"gitops.mxcd.de/secret-id": s.ID,
			},
		},
		Type: v1.SecretType(s.Type),
		StringData: s.Data,
	}, metav1.CreateOptions{})
	log.Trace(k8sSecret)
	if err != nil {
		return err
	}
	println(s.Namespace, "/", s.Name, color.InGreen(" created"))
	return err
}

func UpdateSecret(s *secret.Secret, clusterName string) error {
	if s.Type == "ConfigMap" {
		return UpdateK8sConfigMap(s, clusterName)
	} else {
		return UpdateK8sSecret(s, clusterName)
	}
}

func UpdateK8sSecret(s *secret.Secret, clusterName string) error {
	clusterClient, err := GetClient(clusterName)
	if err != nil {
		return err
	}
	clientset := clusterClient.Clientset

	log.Trace("Updating Secret ", s.Name, " in namespace ", s.Namespace)
	k8sSecret, err := clientset.CoreV1().Secrets(s.Namespace).Update(context.Background(), &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.Name,
			Namespace: s.Namespace,
			Annotations: map[string]string{
				"gitops.mxcd.de/secret-id": s.ID,
			},
		},
		Type: v1.SecretType(s.Type),
		StringData: s.Data,
	}, metav1.UpdateOptions{})
	log.Trace(k8sSecret)
	if err != nil {
		return err
	}
	println(s.Namespace, "/", s.Name, color.InYellow(" updated"))
	return err
}

func UpdateK8sConfigMap(s *secret.Secret, clusterName string) error {
	clusterClient, err := GetClient(clusterName)
	if err != nil {
		return err
	}
	clientset := clusterClient.Clientset
	log.Trace("Updating ConfigMap ", s.Name, " in namespace ", s.Namespace)

	k8sConfigMap, err := clientset.CoreV1().ConfigMaps(s.Namespace).Update(context.Background(), &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.Name,
			Namespace: s.Namespace,
			Annotations: map[string]string{
				"gitops.mxcd.de/secret-id": s.ID,
			},
		},
		Data: s.Data,
	}, metav1.UpdateOptions{})
	log.Trace(k8sConfigMap)
	if err != nil {
		return err
	}
	println(s.Namespace, "/", s.Name, color.InYellow(" updated"))
	return err
}

func DeleteSecret(s *secret.Secret, clusterName string) error {
	if s.Type == "ConfigMap" {
		return DeleteK8sConfigMap(s, clusterName)
	} else {
		return DeleteK8sSecret(s, clusterName)
	}
}

func DeleteK8sConfigMap(s *secret.Secret, clusterName string) error {
	clusterClient, err := GetClient(clusterName)
	if err != nil {
		return err
	}
	clientset := clusterClient.Clientset
	log.Trace("Deleting ConfigMap ", s.Name, " in namespace ", s.Namespace)
	err = clientset.CoreV1().ConfigMaps(s.Namespace).Delete(context.Background(), s.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	println(s.Namespace, "/", s.Name, color.InRed(" deleted"))
	return err
}

func DeleteK8sSecret(s *secret.Secret, clusterName string) error {
	clusterClient, err := GetClient(clusterName)
	if err != nil {
		return err
	}
	clientset := clusterClient.Clientset
	log.Trace("Deleting Secret ", s.Name, " in namespace ", s.Namespace)
	err = clientset.CoreV1().Secrets(s.Namespace).Delete(context.Background(), s.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	println(s.Namespace, "/", s.Name, color.InRed(" deleted"))
	return err
}

func GetSecret(s *secret.Secret, clusterName string) (*secret.Secret, error) {
	if s.Type == "ConfigMap" {
		return GetK8sConfigMap(s, clusterName)
	} else {
		return GetK8sSecret(s, clusterName)
	}
}

func GetK8sConfigMap(s *secret.Secret, clusterName string) (*secret.Secret, error) {
	clusterClient, err := GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	clientset := clusterClient.Clientset
	k8sConfigMap, err := clientset.CoreV1().ConfigMaps(s.Namespace).Get(context.Background(), s.Name, metav1.GetOptions{})
	
	if err != nil {
		return nil, err
	}

	return &secret.Secret{
		Name: k8sConfigMap.Name,
		TargetType: secret.SecretTargetTypeKubernetes,
		Target: clusterName,
		Namespace: k8sConfigMap.Namespace,
		Data: k8sConfigMap.Data,
		Type: "ConfigMap",
	}, nil
}

func GetK8sSecret(s *secret.Secret, clusterName string) (*secret.Secret, error) {
	clusterClient, err := GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	clientset := clusterClient.Clientset
	k8sSecret, err := clientset.CoreV1().Secrets(s.Namespace).Get(context.Background(), s.Name, metav1.GetOptions{})
	
	if err != nil {
		return nil, err
	}

	return &secret.Secret{
		Name: k8sSecret.Name,
		TargetType: secret.SecretTargetTypeKubernetes,
		Target: clusterName,
		Namespace: k8sSecret.Namespace,
		Data: getStringData(k8sSecret),
		Type: string(k8sSecret.Type),
	}, nil
}

func getStringData(s *v1.Secret) map[string]string {
	stringData := make(map[string]string)
	for key, value := range s.Data {
		stringData[key] = string(value)
	}
	return stringData
}