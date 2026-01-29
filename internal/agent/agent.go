package agent

import (
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go/muid"
)

// NewAgent creates a new Agent instance.

func NewAgent(cfg config.AgentConfig, repoRoot api.RepoRoot, bus *EventBus) (*Agent, error) {

	// Generate ID

	id := api.AgentID(muid.Make())

	slug := api.NormalizeAgentSlug(cfg.Name)



	a := &Agent{

		ID:       id,

		Slug:     slug,

		Name:     cfg.Name,

		About:    cfg.About,

		Adapter:  cfg.Adapter,

		RepoRoot: repoRoot,

		Config:   cfg,

		Sessions: make(map[api.SessionID]*Session),

	}



		a.Lifecycle = NewLifecycleHSM(a)



		a.Presence = NewPresenceHSM(a, bus)



	



		GlobalRegistry.Register(a)



	



		return a, nil



	}



	
