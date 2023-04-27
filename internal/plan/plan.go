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
	TargetType secret.SecretTargetType
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

func (p *Plan) NothingToDo() bool {
	for _, item := range p.Items {
		if !item.Diff.Equal {
			return false
		}
	}
	return true
}

func (i *PlanItem) ComputeDiff() {
	i.Diff = secret.CompareSecrets(i.RemoteSecret, i.LocalSecret)
}

func (p *Plan) Print() {
	for i, item := range p.Items {
		item.Diff.Print()
		if i < len(p.Items)-1 {
			println("---")
		}
	}
}

func (p *Plan) Execute() error {
	if p.TargetType == secret.SecretTargetTypeKubernetes {
		return executeKubernetesPlan(p)

	} else if p.TargetType == secret.SecretTargetTypeVault {
		log.Fatal("Not implemented")
		return nil
	}
	return nil
}

func executeKubernetesPlan(p *Plan) error {
	for _, item := range p.Items {
		if item.Diff.Equal {
			log.Trace("Secret ", item.LocalSecret.Namespace, "/", item.LocalSecret.Name, " is equal, skipping...")
			continue
		}
		if item.Diff.Type == secret.SecretDiffTypeAdded {
			log.Trace("Secret ", item.LocalSecret.Namespace, "/", item.LocalSecret.Name, " is new, creating...")
			err := k8s.CreateSecret(item.LocalSecret, item.LocalSecret.Target)
			if err != nil {
				log.Error("Failed to create secret ", item.LocalSecret.Namespace, "/", item.LocalSecret.Name, " in cluster")
				return err
			}
		} else if item.Diff.Type == secret.SecretDiffTypeChanged {
			log.Trace("Secret ", item.LocalSecret.Namespace, "/", item.LocalSecret.Name, " is modified, updating...")
			err := k8s.UpdateSecret(item.LocalSecret, item.LocalSecret.Target)
			if err != nil {
				log.Error("Failed to update secret ", item.LocalSecret.Namespace, "/", item.LocalSecret.Name, " in cluster")
				return err
			}
		} else if item.Diff.Type == secret.SecretDiffTypeRemoved {
			log.Trace("Secret ", item.RemoteSecret.Namespace, "/", item.RemoteSecret.Name, " is deleted, deleting...")
			err := k8s.DeleteSecret(item.RemoteSecret, item.RemoteSecret.Target)
			if err != nil {
				log.Error("Failed to delete secret ", item.RemoteSecret.Namespace, "/", item.RemoteSecret.Name, " in cluster")
				return err
			}
		}
	}
	return nil
}