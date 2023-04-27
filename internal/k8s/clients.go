package k8s

import (
	"fmt"
	"os"
	"regexp"

	"github.com/TwiN/go-color"
	"github.com/mxcd/gitops-cli/internal/state"
	"github.com/mxcd/gitops-cli/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type ClusterClient struct {
	Name string
	Clientset *kubernetes.Clientset
	Connected bool
	ClusterVersion string
}

var clusterClients map[string] *ClusterClient = make(map[string] *ClusterClient)

func InitClusterClients(c *cli.Context) error {
	log.Trace("Initializing cluster clients")

	stateClusters := state.GetState().GetClusters()
	err := InitDefaultClusterClient(c)
	if err != nil && len(stateClusters) == 0 {
		log.Error("No default cluster config available")
		return err
	}
	log.Trace("Found clusters in state. Initializing cluster clients")
	for _, cluster := range stateClusters {
		log.Trace("Initializing cluster client for cluster: ", cluster.Name)
		AddClient(cluster.Name, cluster.ConfigFile)
	}
	return nil
}

func InitDefaultClusterClient(c *cli.Context) error {
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
		log.Warn("No default cluster config available")
		return nil
	}

	log.Trace("Creating Kubernetes clientset")
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	
	clusterClient := &ClusterClient{
		Name: string(util.DefaultClusterClient),
		Clientset: clientset,
		Connected: false,
	}
	clusterClient.TestConnection()
	clusterClients[string(util.DefaultClusterClient)] = clusterClient

	return nil
}

func GetClient(clusterName string) (*ClusterClient, error) {
	if clusterClients[clusterName] == nil {
		return nil, fmt.Errorf("client '%s' not found", clusterName)
	}
	return clusterClients[clusterName], nil
}

func GetClients() map[string] *ClusterClient {
	return clusterClients
}

func AddClient(clusterName string, kubeconfigFile string) error {
	log.Trace("Building Kubernetes client config from KUBECONFIG: ", kubeconfigFile)
	// check if the file exists
	_, err := os.Stat(kubeconfigFile)
	if os.IsNotExist(err) {
		log.Error("KUBECONFIG file not found: ", kubeconfigFile)
		return err
	}

	secretKubeconfigRegex, err := regexp.Compile(`.*\.kubeconfig\.secret\.enc\.ya?ml$`)
	if err != nil {
		log.Fatal(err)
	}

	var kubeconfigFileData []byte = []byte{}

	if secretKubeconfigRegex.MatchString(kubeconfigFile) {
		log.Trace("KUBECONFIG file is a secret. Decrypting")
		kubeconfigFileData, err = util.DecryptFile(kubeconfigFile)
		if err != nil {
			log.Error("Failed to decrypt KUBECONFIG file: ", err)
			return err
		}
	} else {
		log.Trace("KUBECONFIG file is not a secret. Reading")
		kubeconfigFileData, err = os.ReadFile(kubeconfigFile)
		if err != nil {
			log.Error("Failed to read KUBECONFIG file: ", err)
			return err
		}
	}

	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfigFileData)
	if err != nil {
		log.Error("Failed to build Kubernetes client config: ", err)
		return err
	}
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		log.Error("Failed to build Kubernetes client config: ", err)
		return err
	}
	
	log.Trace("Creating Kubernetes clientset")
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	clusterClient := &ClusterClient{
		Name: clusterName,
		Clientset: clientset,
		Connected: false,
	}
	clusterClient.TestConnection()
	clusterClients[clusterName] = clusterClient
	return nil
}

func (c *ClusterClient) TestConnection() (bool, error) {
	log.Trace("Testing Kubernetes cluster connection")
	serverVersion, err := c.Clientset.Discovery().ServerVersion()
	if err != nil {
		log.Warn("Failed to connect to Kubernetes cluster ", color.InBlue(c.Name), ": ", err)
		c.Connected = false
		return false, err
	}
	log.Debug("Connected to Kubernetes cluster: ", serverVersion)
	c.Connected = true
	c.ClusterVersion = serverVersion.String()
	return true, nil
}

func (c *ClusterClient) PrettyPrint() {
	clusterInfoLine := fmt.Sprintf("Cluster: %s (%s)\t", color.InBlue(c.Name), color.InGray(c.ClusterVersion))
	if c.Connected {
		clusterInfoLine = fmt.Sprintf("%sConnected: %s", clusterInfoLine, color.InGreen("true"))
	} else {
		clusterInfoLine = fmt.Sprintf("%sConnected: %s", clusterInfoLine, color.InRed("false"))
	}
	println(clusterInfoLine)
}