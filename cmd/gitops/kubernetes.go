package main

import (
	"fmt"
	"strings"

	"github.com/TwiN/go-color"
	"github.com/google/uuid"
	"github.com/mxcd/gitops-cli/internal/k8s"
	"github.com/mxcd/gitops-cli/internal/plan"
	"github.com/mxcd/gitops-cli/internal/secret"
	"github.com/mxcd/gitops-cli/internal/state"
	"github.com/mxcd/gitops-cli/internal/util"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

func applyKubernetes(c *cli.Context) error {
	p, err := createKubernetesPlan(c)
	if err != nil {
		return err
	}

	// exit if there is nothing to do
	if p.NothingToDo() {
		println(color.InGreen("No changes to apply."))
		exitApplication(c, true)
		return nil
	}

	prettyPrintPlan(p, c.Bool("show-unchanged"))

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

	exitApplication(c, true)
	return nil
}

func planKubernetes(c *cli.Context) error {
	clusterLimitString := getClusterLimit(c)

	p, err := createKubernetesPlan(c)
	if err != nil {
		return err
	}

	if p.NothingToDo() {
		println(color.InGreen("No changes to apply."))
	} else {
		prettyPrintPlan(p, c.Bool("show-unchanged"))
		dirLimitString := ""
		if c.String("dir") != "" {
			dirLimitString = " --dir " + c.String("dir")
		}
		applyString := fmt.Sprintf("gitops secrets%s apply kubernetes %s", dirLimitString, clusterLimitString)
		println(color.InBold("use"), color.InGreen(color.InBold(applyString)), color.InBold("to apply these changes to your cluster"))
	}
	exitApplication(c, false)
	return nil
}

func createKubernetesPlan(c *cli.Context) (*plan.Plan, error) {
	log.Trace("Creating Kubernetes plan")
	clusterLimit := getClusterLimit(c)

	limitPrintln := false

	if clusterLimit != "" {
		println("Limiting to cluster " + color.InBlue(clusterLimit))
		limitPrintln = true
	}

	dirLimit := getDirLimit(c)
	if dirLimit != "" {
		println("Limiting to directory " + color.InPurple(dirLimit))
		limitPrintln = true
	}

	if limitPrintln {
		println("")
	}

	err := k8s.InitClusterClients(c)
	if err != nil {
		log.Error("Failed to init Kubernetes cluster connection")
		return nil, err
	}

	localSecrets, err := secret.LoadLocalSecretsLimited(secret.SecretTargetTypeKubernetes, dirLimit, clusterLimit)
	if err != nil {
		log.Error("Failed to load local secrets with target ", secret.SecretTargetTypeKubernetes)
		return nil, err
	}
	log.Trace("Loaded ", len(localSecrets), " local secrets with target ", secret.SecretTargetTypeKubernetes)

	p := &plan.Plan{
		TargetType: secret.SecretTargetTypeKubernetes,
		Items:      []plan.PlanItem{},
	}

	bar := progressbar.NewOptions(len(localSecrets),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionSetElapsedTime(false),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetDescription("[green][Syncing local state with cluster][reset]"),
	)

	for _, localSecret := range localSecrets {
		bar.Add(1)
		// check for local secret in state
		// update ID in localSecret if secret exists in state
		// update hash in state if secret exists in state
		stateSecret := state.GetState().GetByPath(localSecret.Path)
		if stateSecret == nil {
			log.Trace("Secret ", localSecret.CombinedName(), " does not exist in state")
			localSecret.ID = uuid.New().String()
			stateSecret = state.GetState().Add(localSecret)
		} else {
			log.Trace("Secret ", localSecret.CombinedName(), " exists in state. Updating")
			stateSecret.Update(localSecret)
		}

		planItem := plan.PlanItem{
			LocalSecret: localSecret,
		}

		remoteSecret, err := k8s.GetSecret(localSecret, localSecret.Target)
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
	bar.Finish()
	println("")
	println("")

	updatedStateSecrets := state.GetState().Secrets[:0]
	for _, stateSecret := range state.GetState().Secrets {
		stateSecretFound := false
		for _, localSecret := range localSecrets {
			if stateSecret.Path == localSecret.Path {
				stateSecretFound = true
				break
			}
		}

		// only add secret to updated state, if it still exists locally or if it is not in the scope of the current dir or cluster limits
		if stateSecretFound || !strings.HasPrefix(stateSecret.Path, dirLimit) || (stateSecret.Target != clusterLimit && clusterLimit != "") {
			updatedStateSecrets = append(updatedStateSecrets, stateSecret)
			continue
		}

		// at this point, the local secret does not exist anymore, but the secret is still in the state
		// therefore we are checking if the cluster secret actually exists
		log.Trace("Secret ", stateSecret.CombinedName(), " does not exist locally")
		remoteSecret, err := k8s.GetSecret(&secret.Secret{
			Name:      stateSecret.Name,
			Namespace: stateSecret.Namespace,
			Type:      stateSecret.Type,
		}, stateSecret.Target)
		if err != nil {
			// only throw error if err is not "not found"
			if !k8sErrors.IsNotFound(err) {
				log.Error("Failed to get state secret ", stateSecret.CombinedName(), " from Kubernetes cluster")
				return nil, err
			}
		}

		// at this point, the local secret does not exist anymore, but the secret is still in the state
		// now we are checking if the cluster secret actually exists
		if remoteSecret == nil {
			log.Trace("State secret ", stateSecret.CombinedName(), " does not exist in Kubernetes cluster")
			// do not add secret to updated state secrets
			continue
		}

		// at this state, the local secret does not exist anymore, but the secret is still in the state
		// also, the cluster still holds the secret which needs to be deleted
		planItem := plan.PlanItem{
			LocalSecret:  nil,
			RemoteSecret: remoteSecret,
		}
		planItem.ComputeDiff()
		p.AddItem(planItem)
	}

	// update state secrets
	state.GetState().SetSecrets(updatedStateSecrets)
	return p, nil
}

func getClusterLimit(c *cli.Context) string {
	clusterLimit := c.Args().Get(0)
	if clusterLimit != "" {
		log.Trace("Limiting to cluster ", clusterLimit)
	} else {
		log.Trace("No cluster limit set. Applying to all clusters.")
	}
	return clusterLimit
}

func getDirLimit(c *cli.Context) string {
	dirLimit := c.String("dir")
	if dirLimit != "" {
		log.Trace("Limiting to directory ", dirLimit)
	} else {
		log.Trace("No dir limit set. Collecting secrets in all directories.")
	}
	return dirLimit
}

func prettyPrintPlan(p *plan.Plan, showUnchanged bool) {
	println("")
	println("GitOps CLI computed the following changes for your cluster:")
	println("-------------------------------------------------------")
	println("")
	p.Print(showUnchanged)
	println("")
	println("-------------------------------------------------------")
	println("")
}
