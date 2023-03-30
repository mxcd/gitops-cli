package k8s

import (
	"context"
	"os"

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
	log.Info("Connected to Kubernetes cluster: ", serverVersion)

	return true, nil
}

func CreateSecret(s *secret.Secret) error {
	log.Trace("Creating secret ", s.Name, " in namespace ", s.Namespace)
	k8sSecret, err := clientset.CoreV1().Secrets(s.Namespace).Create(context.Background(), &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.Name,
			Namespace: s.Namespace,
		},
		Type: v1.SecretTypeOpaque,
		StringData: s.Data,
	}, metav1.CreateOptions{})
	log.Trace(k8sSecret)
	if err != nil {
		return err
	}
	log.Info(s.Namespace, "/", s.Name, " created")
	return err
}

func UpdateSecret(s *secret.Secret) error {
	log.Trace("Updating secret ", s.Name, " in namespace ", s.Namespace)
	k8sSecret, err := clientset.CoreV1().Secrets(s.Namespace).Update(context.Background(), &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.Name,
			Namespace: s.Namespace,
		},
		Type: v1.SecretTypeOpaque,
		StringData: s.Data,
	}, metav1.UpdateOptions{})
	log.Trace(k8sSecret)
	if err != nil {
		return err
	}
	log.Info(s.Namespace, "/", s.Name, " updated")
	return err
}

func DeleteSecret(s *secret.Secret) error {
	log.Trace("Deleting secret ", s.Name, " in namespace ", s.Namespace)
	err := clientset.CoreV1().Secrets(s.Namespace).Delete(context.Background(), s.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	log.Info(s.Namespace, "/", s.Name, " deleted")
	return err
}

func GetSecret(s *secret.Secret) (*secret.Secret, error) {
	k8sSecret, err := clientset.CoreV1().Secrets(s.Namespace).Get(context.Background(), s.Name, metav1.GetOptions{})
	
	if err != nil {
		return nil, err
	}

	return &secret.Secret{
		Name: k8sSecret.Name,
		Target: secret.SecretTargetKubernetes,
		Namespace: k8sSecret.Namespace,
		Data: k8sSecret.StringData,
		Type: string(k8sSecret.Type),
	}, nil
}