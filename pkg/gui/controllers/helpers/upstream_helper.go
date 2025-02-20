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
	PromptForUpstreamBranch(chosenRemote string, initialBranch string, onConfirm func(Upstream) error) error
	PromptForUpstream(initialContent Upstream, onConfirm func(Upstream) error) error
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

func (self *UpstreamHelper) PromptForUpstreamBranch(chosenRemote string, initialBranch string, onConfirm func(Upstream) error) error {
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

func (self *UpstreamHelper) PromptForUpstream(initialContent Upstream, onConfirm func(Upstream) error) error {
	self.c.Prompt(types.PromptOpts{
		Title:               self.c.Tr.SelectTargetRemote,
		InitialContent:      initialContent.Remote,
		FindSuggestionsFunc: self.suggestions.GetRemoteSuggestionsFunc(),
		HandleConfirm: func(toRemote string) error {
			return self.PromptForUpstreamBranch(toRemote, initialContent.Branch, onConfirm)
		},
	})

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
