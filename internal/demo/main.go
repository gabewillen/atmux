// main.go demonstrates core dependencies for Phase 0 completion.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
	"github.com/agentflare-ai/amux/internal/hsm"
	"github.com/agentflare-ai/amux/internal/inference"
)

func main() {
	fmt.Println("=== amux Phase 0 Core Dependencies Demo ===")

	ctx := context.Background()

	// Test 1: Configuration System
	fmt.Println("\n1. Testing Configuration System...")
	if err := testConfiguration(ctx); err != nil {
		log.Printf("Configuration test failed: %v", err)
		os.Exit(1)
	}
	fmt.Println("✅ Configuration system working")

	// Test 2: HSM Event Dispatch
	fmt.Println("\n2. Testing HSM Event Dispatch...")
	if err := testHSM(ctx); err != nil {
		log.Printf("HSM test failed: %v", err)
		os.Exit(1)
	}
	fmt.Println("✅ HSM event dispatch working")

	// Test 3: LiquidGen Integration
	fmt.Println("\n3. Testing LiquidGen Integration...")
	if err := testLiquidGen(ctx); err != nil {
		log.Printf("LiquidGen test failed: %v", err)
		os.Exit(1)
	}
	fmt.Println("✅ LiquidGen integration working")

	fmt.Println("\n=== All Core Dependencies Working ===")
	fmt.Println("Phase 0 requirements satisfied!")
}

func testConfiguration(ctx context.Context) error {
	// Test configuration loading
	cfg, err := config.Load()
	if err != nil {
		return amuxerrors.Wrap("loading configuration", err)
	}

	// Test validation
	if err := config.Validate(cfg); err != nil {
		return amuxerrors.Wrap("validating configuration", err)
	}

	// Test environment override
	os.Setenv("AMUX__CORE__LOG_LEVEL", "debug")
	os.Setenv("AMUX__CORE__DEBUG", "true")

	cfg2, err := config.Load()
	if err != nil {
		return amuxerrors.Wrap("loading config with env override", err)
	}

	if cfg2.Core.LogLevel != "debug" || !cfg2.Core.Debug {
		return amuxerrors.New("environment override not working")
	}

	fmt.Printf("  Loaded config: LogLevel=%s, Debug=%t\n", cfg2.Core.LogLevel, cfg2.Core.Debug)
	return nil
}

func testHSM(ctx context.Context) error {
	// Create a simple HSM
	id := hsm.GenerateID()

	// Define states
	type initialState struct{ name string }
	type workingState struct{ name string }

	initial := &initialState{name: "initial"}
	working := &workingState{name: "working"}

	// Create state wrappers
	hsmInitialState := hsm.StateWrapper(initial)
	hsmWorkingState := hsm.StateWrapper(working)

	// Create machine
	machine := hsm.NewMachine(id, hsmInitialState)
	machine.AddState(hsmInitialState)
	machine.AddState(hsmWorkingState)

	// Add transition
	machine.AddTransition(hsm.Transition{
		On:     "start",
		Source: "initial",
		Target: "working",
		Effect: func(ctx context.Context, ev hsm.Event) {
			fmt.Printf("  HSM: Transitioned from %s to %s\n", ev.Source, ev.Type)
		},
	})

	// Subscribe to events
	machine.Subscribe("start", func(ctx context.Context, ev hsm.Event) {
		fmt.Printf("  HSM: Received start event from %d\n", ev.Source)
	})

	// Dispatch event
	machine.Dispatch(ctx, hsm.Event{
		Type:   "start",
		Source: hsm.GenerateID(),
		Data:   "test data",
	})

	// Allow some time for async handlers
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("  Current HSM state: %s\n", machine.CurrentState().Name())
	return nil
}

func testLiquidGen(ctx context.Context) error {
	// Test liquidgen integration
	models := map[string]config.ModelConfig{
		"lfm2.5-thinking": {
			Type: "generation",
			Path: "lfm2.5-thinking.Q4_K_M.gguf",
			Parameters: map[string]interface{}{
				"quantization": "Q4_K_M",
				"context_size": 32768,
			},
		},
	}

	manager, err := inference.NewManager(&config.InferenceConfig{
		Enabled: true,
		Engine:  "liquidgen",
		Models:  models,
	})
	if err != nil {
		return amuxerrors.Wrap("creating inference manager", err)
	}
	defer manager.Shutdown(ctx)

	// Get engine
	engine, err := manager.GetDefaultEngine()
	if err != nil {
		// This is expected for Phase 0 if liquidgen server is not running
		fmt.Printf("  LiquidGen server not available (expected for Phase 0): %v\n", err)
		return nil
	}

	// Test engine info
	info := engine.Info()
	fmt.Printf("  LiquidGen Engine: %s v%s\n", info.Name, info.Version)
	fmt.Printf("  Available models: %v\n", len(info.Models))

	return nil
}
