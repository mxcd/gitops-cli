package main

import (
	"github.com/TwiN/go-color"
	"github.com/mxcd/gitops-cli/internal/k8s"
	"github.com/mxcd/gitops-cli/internal/plan"
	"github.com/mxcd/gitops-cli/internal/secret"
	"github.com/mxcd/gitops-cli/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

func applyKubernetes(c *cli.Context) error {
	p, err := createKubernetesPlan(c)
	if err != nil {
		return err
	}

	prettyPrintPlan(p)

	if p.NothingToDo() {
		println(color.InGreen("No changes to apply."))
		return nil
	}
	
	if !c.Bool("auto-approve") {
		println("GitOps CLI will apply these changes to your Kubernetes cluster.")
		println("Only 'yes' will be accepted to approve.")
		promtAnswer := util.StringPrompt("Apply changes above: ")
		
		if promtAnswer != "yes" {
			println("Aborting")
			return nil
		}
	}

	println("GitOps CLI will now execute the changes for your cluster:")
	println("-------------------------------------------------------")
	println("")
	err = p.Execute()
	if err != nil {
		return err
	}
	println("")
	println("-------------------------------------------------------")
	println("")
	println(color.InGreen("All changes applied."))
	return nil
}

func planKubernetes(c *cli.Context) error {
	p, err := createKubernetesPlan(c)
	if err != nil {
		return err
	}
	prettyPrintPlan(p)

	if p.NothingToDo() {
		println(color.InGreen("No changes to apply."))
	} else {
		println(color.InBold("use"), color.InGreen(color.InBold("gitops secrets apply kubernetes")), color.InBold("to apply these changes to your cluster"))
	}
	return nil
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

	return p, nil
}

func prettyPrintPlan(p *plan.Plan) {
	println("GitOps CLI computed the following changes for your cluster:")
	println("-------------------------------------------------------")
	println("")
	p.Print()
	println("")
	println("-------------------------------------------------------")
	println("")
}