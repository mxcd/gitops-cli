package k8s

import (
	"context"
	"os"

	"github.com/TwiN/go-color"
	"github.com/mxcd/gitops-cli/internal/secret"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var clientset *kubernetes.Clientset

func InitKubernetesClient(c *cli.Context) error {
	kubeconfig := c.String("kubeconfig")
	if kubeconfig == "" {
		log.Trace("No KUBECONFIG given. Testing default locations")
		// check if file exists
		_, err := os.Stat(clientcmd.RecommendedHomeFile)
		if os.IsNotExist(err) {
			log.Trace("No KUBECONFIG found in default locations. Attempting in-cluster config")
		} else if err == nil {
			log.Trace("Found KUBECONFIG in default location: ", clientcmd.RecommendedHomeFile)
			kubeconfig = clientcmd.RecommendedHomeFile
		}
	}

	log.Trace("Building Kubernetes client config from KUBECONFIG: ", kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	log.Trace("Creating Kubernetes clientset")
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	return nil
}

func TestClusterConnection() (bool, error) {
	log.Trace("Testing Kubernetes cluster connection")
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		log.Error("Failed to connect to Kubernetes cluster: ", err)
		return false, err
	}
	log.Debug("Connected to Kubernetes cluster: ", serverVersion)

	return true, nil
}

func CreateSecret(s *secret.Secret) error {
	if s.Type == "ConfigMap" {
		return CreateK8sConfigMap(s)
	} else {
		return CreateK8sSecret(s)
	}
}

func CreateK8sConfigMap(s *secret.Secret) error {
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

func CreateK8sSecret(s *secret.Secret) error {
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

func UpdateSecret(s *secret.Secret) error {
	if s.Type == "ConfigMap" {
		return UpdateK8sConfigMap(s)
	} else {
		return UpdateK8sSecret(s)
	}
}

func UpdateK8sSecret(s *secret.Secret) error {
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

func UpdateK8sConfigMap(s *secret.Secret) error {
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

func DeleteSecret(s *secret.Secret) error {
	if s.Type == "ConfigMap" {
		return DeleteK8sConfigMap(s)
	} else {
		return DeleteK8sSecret(s)
	}
}

func DeleteK8sConfigMap(s *secret.Secret) error {
	log.Trace("Deleting ConfigMap ", s.Name, " in namespace ", s.Namespace)
	err := clientset.CoreV1().ConfigMaps(s.Namespace).Delete(context.Background(), s.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	println(s.Namespace, "/", s.Name, color.InRed(" deleted"))
	return err
}

func DeleteK8sSecret(s *secret.Secret) error {
	log.Trace("Deleting Secret ", s.Name, " in namespace ", s.Namespace)
	err := clientset.CoreV1().Secrets(s.Namespace).Delete(context.Background(), s.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	println(s.Namespace, "/", s.Name, color.InRed(" deleted"))
	return err
}

func GetSecret(s *secret.Secret) (*secret.Secret, error) {
	if s.Type == "ConfigMap" {
		return GetK8sConfigMap(s)
	} else {
		return GetK8sSecret(s)
	}
}

func GetK8sConfigMap(s *secret.Secret) (*secret.Secret, error) {
	k8sConfigMap, err := clientset.CoreV1().ConfigMaps(s.Namespace).Get(context.Background(), s.Name, metav1.GetOptions{})
	
	if err != nil {
		return nil, err
	}

	return &secret.Secret{
		Name: k8sConfigMap.Name,
		Target: secret.SecretTargetKubernetes,
		Namespace: k8sConfigMap.Namespace,
		Data: k8sConfigMap.Data,
		Type: "ConfigMap",
	}, nil
}

func GetK8sSecret(s *secret.Secret) (*secret.Secret, error) {
	k8sSecret, err := clientset.CoreV1().Secrets(s.Namespace).Get(context.Background(), s.Name, metav1.GetOptions{})
	
	if err != nil {
		return nil, err
	}

	return &secret.Secret{
		Name: k8sSecret.Name,
		Target: secret.SecretTargetKubernetes,
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