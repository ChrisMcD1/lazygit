package helpers

import (
	"fmt"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/samber/lo"
)

type UpstreamHelper struct {
	c *HelperCommon

	suggestions *SuggestionsHelper
}

type IUpstreamHelper interface {
	PromptForUpstream(suggestedBranch string, onConfirm func(Upstream) error) error
}

var _ IUpstreamHelper = &UpstreamHelper{}

func NewUpstreamHelper(
	c *HelperCommon,
	suggestions *SuggestionsHelper,
) *UpstreamHelper {
	return &UpstreamHelper{
		c:           c,
		suggestions: suggestions,
	}
}

type Upstream struct {
	Remote string
	Branch string
}

func (self *UpstreamHelper) promptForUpstreamBranch(chosenRemote string, initialBranch string, onConfirm func(Upstream) error) error {
	self.c.Log.Debugf("User selected remote '%s'", chosenRemote)
	remoteDoesNotExist := lo.NoneBy(self.c.Model().Remotes, func(remote *models.Remote) bool {
		return remote.Name == chosenRemote
	})
	if remoteDoesNotExist {
		return fmt.Errorf(self.c.Tr.NoValidRemoteName, chosenRemote)
	}

	self.c.Prompt(types.PromptOpts{
		Title:               fmt.Sprintf("Targeting remote %s", chosenRemote),
		InitialContent:      initialBranch,
		FindSuggestionsFunc: self.suggestions.GetRemoteBranchesForRemoteSuggestionsFunc(chosenRemote),
		HandleConfirm: func(chosenBranch string) error {
			self.c.Log.Debugf("User selected branch '%s' on remote '%s'", chosenRemote, chosenBranch)
			return onConfirm(Upstream{chosenRemote, chosenRemote})
		},
	})
	return nil
}

func (self *UpstreamHelper) PromptForUpstream(suggestedBranch string, onConfirm func(Upstream) error) error {
	if len(self.c.Model().Remotes) == 1 {
		remote := self.c.Model().Remotes[0].Name
		self.c.Log.Debugf("Defaulting to only remote %s", remote)
		return self.promptForUpstreamBranch(remote, suggestedBranch, onConfirm)
	} else {
		suggestedRemote := getSuggestedRemote(self.c.Model().Remotes)
		self.c.Prompt(types.PromptOpts{
			Title:               self.c.Tr.SelectTargetRemote,
			InitialContent:      suggestedRemote,
			FindSuggestionsFunc: self.suggestions.GetRemoteSuggestionsFunc(),
			HandleConfirm: func(toRemote string) error {
				return self.promptForUpstreamBranch(toRemote, suggestedBranch, onConfirm)
			},
		})
	}

	return nil
}

func getSuggestedRemote(remotes []*models.Remote) string {
	if len(remotes) == 0 {
		return "origin"
	}

	for _, remote := range remotes {
		if remote.Name == "origin" {
			return remote.Name
		}
	}

	return remotes[0].Name
}
