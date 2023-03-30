package plan

import (
	log "github.com/sirupsen/logrus"

	"github.com/mxcd/gitops-cli/internal/k8s"
	"github.com/mxcd/gitops-cli/internal/secret"
)

type Plan struct {
	// List of items in the plan
	Items []PlanItem
	// Target type of the plan
	Target secret.SecretTarget
}

type PlanItem struct {
	// Pointer to the local secret of this plan item
	LocalSecret *secret.Secret
	// Pointer to the remote secret of this plan item
	RemoteSecret *secret.Secret
	// Pointer to the diff between the local and remote secret
	Diff *secret.SecretDiff
}

func (p *Plan) AddItem(item PlanItem) {
	p.Items = append(p.Items, item)
}

func (i *PlanItem) ComputeDiff() {
	i.Diff = secret.CompareSecrets(i.RemoteSecret, i.LocalSecret)
}

func (p *Plan) Print() {
	for _, item := range p.Items {
		item.Diff.Print()
	}
}

func (p *Plan) Execute() {
	if p.Target == secret.SecretTargetKubernetes {
		executeKubernetesPlan(p)
	} else if p.Target == secret.SecretTargetVault {
		log.Fatal("Not implemented")
	}
}

func executeKubernetesPlan(p *Plan) {
	for _, item := range p.Items {
		if item.Diff.Equal {
			log.Trace("Secret ", item.LocalSecret.Name, " is equal, skipping")
			continue
		}
		if item.Diff.Type == secret.SecretDiffTypeAdded {
			log.Trace("Secret ", item.LocalSecret.Name, " is new, creating")
			k8s.CreateSecret(item.LocalSecret)
		} else if item.Diff.Type == secret.SecretDiffTypeChanged {
			log.Trace("Secret ", item.LocalSecret.Name, " is modified, updating")
			k8s.UpdateSecret(item.LocalSecret)
		} else if item.Diff.Type == secret.SecretDiffTypeRemoved {
			log.Trace("Secret ", item.LocalSecret.Name, " is deleted, deleting")
			k8s.DeleteSecret(item.LocalSecret)
		}
	}
}