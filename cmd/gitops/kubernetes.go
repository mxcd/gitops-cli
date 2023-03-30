package main

import (
	"github.com/mxcd/gitops-cli/internal/k8s"
	"github.com/mxcd/gitops-cli/internal/plan"
	"github.com/mxcd/gitops-cli/internal/secret"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

func applyKubernetes(c *cli.Context) error {
	p, err := createKubernetesPlan(c)
	if err != nil {
		return err
	}
	p.Execute()
	return nil
}

func planKubernetes(c *cli.Context) error {
	_, err := createKubernetesPlan(c)
	return err
}

func createKubernetesPlan(c *cli.Context) (*plan.Plan, error) {
	err := k8s.InitKubernetesClient(c)
	if err != nil {
		log.Error("Failed to init Kubernetes cluster connection")
		return nil, err
	}
	connectionEstablished, err := k8s.TestClusterConnection()
	if !connectionEstablished || err != nil {
		log.Error("Failed to connect to Kubernetes cluster")
		return nil, err
	}
	localSecrets, err := secret.LoadLocalSecrets(c.String("root-dir"), secret.SecretTargetKubernetes)
	if err != nil {
		log.Error("Failed to load local secrets with target ", secret.SecretTargetKubernetes)
		return nil, err
	}
	log.Trace("Loaded ", len(localSecrets), " local secrets with target ", secret.SecretTargetKubernetes)
	
	p := &plan.Plan{
		Target: secret.SecretTargetKubernetes,
		Items: []plan.PlanItem{},
	}

	for _, localSecret := range localSecrets {
		planItem := plan.PlanItem{
			LocalSecret: localSecret,
		}
		remoteSecret, err := k8s.GetSecret(localSecret)
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				log.Trace("Secret ", localSecret.Name, " does not exist in Kubernetes cluster")
			} else {
				log.Error("Failed to get secret ", localSecret.Name, " from Kubernetes cluster")
				return nil, err
			}
		}

		planItem.RemoteSecret = remoteSecret
		planItem.ComputeDiff()
		p.AddItem(planItem)
	}

	p.Print()

	return p, nil
}