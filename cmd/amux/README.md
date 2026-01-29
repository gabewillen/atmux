- `adapterName`
- `agentAddCmd`
- `agentCmd`
- `agentListCmd`
- `agentMergeCmd`
- `agentRemoveCmd`
- `agentStartCmd`
- `func findProjectRoot(startDir string) string`
- `func init()`
- `func main()`
- `regressionFlag, noSnapshotFlag`
- `rootCmd`
- `testCmd`
- `type TestSnapshot` — TestSnapshot represents the verification snapshot.

### Variables

#### regressionFlag, noSnapshotFlag

```go
var (
	regressionFlag bool
	noSnapshotFlag bool
)
```

#### adapterName

```go
var adapterName string
```

#### agentAddCmd

```go
var agentAddCmd = &cobra.Command{
	Use:   "add <name> <repo_root>",
	Short: "Add a new agent",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		repoPath := args[1]

		resolver, err := paths.NewResolver()
		if err != nil {
			return err
		}

		repoRoot, err := resolver.CanonicalizeRepoRoot(repoPath)
		if err != nil {
			return errors.Wrapf(err, "invalid repository path %q", repoPath)
		}

		if !worktree.IsRepo(repoRoot) {
			return fmt.Errorf("path %s is not a git repository", repoRoot)
		}

		def := config.AgentDef{
			Name:    name,
			Adapter: adapterName,
			Location: config.AgentLocation{
				Type:     "local",
				RepoPath: repoRoot,
			},
		}

		cwd, _ := os.Getwd()
		projectRoot := findProjectRoot(cwd)

		targetConfigRoot := ""
		if projectRoot != "" {
			targetConfigRoot = projectRoot
			cmd.Printf("Adding agent to project config: %s\n", projectRoot)
		} else {
			cmd.Printf("Adding agent to user config (~/.config/amux)\n")
		}

		if err := config.AddAgent(targetConfigRoot, def); err != nil {
			return err
		}

		cmd.Printf("Agent %q added successfully.\n", name)
		return nil
	},
}
```

#### agentCmd

```go
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage agents",
}
```

#### agentListCmd

```go
var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		projectRoot := findProjectRoot(cwd)

		cfg, err := config.Load(projectRoot)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), "NAME\tADAPTER\tLOCATION")
		for _, a := range cfg.Agents {
			loc := a.Location.RepoPath
			if loc == "" {
				loc = a.Location.Host
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", a.Name, a.Adapter, loc)
		}
		return nil
	},
}
```

#### agentMergeCmd

```go
var agentMergeCmd = &cobra.Command{
	Use:   "merge <name>",
	Short: "Merge agent worktree back to base branch",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cwd, _ := os.Getwd()
		projectRoot := findProjectRoot(cwd)
		cfg, err := config.Load(projectRoot)
		if err != nil {
			return err
		}

		var agentDef *config.AgentDef
		for _, a := range cfg.Agents {
			if a.Name == name {
				agentDef = &a
				break
			}
		}
		if agentDef == nil {
			return fmt.Errorf("agent %q not found", name)
		}

		mgr, err := worktree.NewManager()
		if err != nil {
			return err
		}

		slug := api.NewAgentSlug(name)
		agentData := api.Agent{
			Name:     name,
			Slug:     slug,
			RepoRoot: agentDef.Location.RepoPath,
		}

		strategy := cfg.Git.Merge.Strategy
		if strategy == "" {
			strategy = "squash"
		}
		allowDirty := cfg.Git.Merge.AllowDirty

		cmd.Printf("Merging agent %q using strategy %q (allow_dirty=%v)...\n", name, strategy, allowDirty)
		if err := mgr.MergeAgent(agentData, strategy, allowDirty); err != nil {
			return err
		}

		cmd.Printf("Successfully merged agent %q.\n", name)
		return nil
	},
}
```

#### agentRemoveCmd

```go
var agentRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cwd, _ := os.Getwd()
		projectRoot := findProjectRoot(cwd)

		removed := false
		if projectRoot != "" {
			if err := config.RemoveAgent(projectRoot, name); err == nil {
				cmd.Printf("Removed from project config.\n")
				removed = true
			}
		}

		if !removed {
			if err := config.RemoveAgent("", name); err == nil {
				cmd.Printf("Removed from user config.\n")
				removed = true
			}
		}

		if !removed {
			return fmt.Errorf("agent %q not found", name)
		}
		return nil
	},
}
```

#### agentStartCmd

```go
var agentStartCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start an agent (creates worktree)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cwd, _ := os.Getwd()
		projectRoot := findProjectRoot(cwd)
		cfg, err := config.Load(projectRoot)
		if err != nil {
			return err
		}

		var agentDef *config.AgentDef
		for _, a := range cfg.Agents {
			if a.Name == name {
				agentDef = &a
				break
			}
		}
		if agentDef == nil {
			return fmt.Errorf("agent %q not found", name)
		}

		mgr, err := worktree.NewManager()
		if err != nil {
			return err
		}

		slug := api.NewAgentSlug(name)
		agentData := api.Agent{
			Name:     name,
			Slug:     slug,
			RepoRoot: agentDef.Location.RepoPath,
		}

		wtPath, err := mgr.Ensure(agentData)
		if err != nil {
			return err
		}

		a := agent.NewAgent(name, agentDef.Adapter, agentDef.Location.RepoPath, mgr)

		a.Start()

		cmd.Printf("Waiting for agent to start...\n")

		for i := 0; i < 20; i++ {
			time.Sleep(100 * time.Millisecond)
			if a.PtyFile() != nil {
				break
			}
		}

		ptmx := a.PtyFile()
		if ptmx == nil {
			return fmt.Errorf("agent failed to start (no PTY)")
		}

		cmd.Printf("Agent %q started (worktree: %s)\n", name, wtPath)
		cmd.Printf("Attached. Press Ctrl+C to exit.\n")

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

		go func() {
			io.Copy(os.Stdout, ptmx)
		}()
		go func() {
			io.Copy(ptmx, os.Stdin)
		}()

		<-sigCh
		cmd.Printf("\nStopping agent...\n")
		a.Stop()
		time.Sleep(500 * time.Millisecond)
		return nil
	},
}
```

#### rootCmd

```go
var rootCmd = &cobra.Command{
	Use:   "amux",
	Short: "Agent Multiplexer CLI",
	Long:  `amux is an agent-agnostic orchestrator for coding agents.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		_, err := config.Load("")
		if err != nil {

		}
		return nil
	},
}
```

#### testCmd

```go
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run verification suite and snapshot",
	Long:  `Runs the verification suite and manages regression snapshots.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resolver, err := paths.NewResolver()
		if err != nil {
			return err
		}
		fmt.Printf("Using config dir: %s\n", resolver.ConfigDir())

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		snapshotsDir := filepath.Join(cwd, "snapshots")
		if err := paths.EnsureDir(snapshotsDir); err != nil {
			return err
		}

		currentSnapshot := TestSnapshot{
			Timestamp: time.Now().UTC(),
			Version:   "0.0.0-phase1",
			Phase:     "1",
		}

		if regressionFlag {

			entries, err := os.ReadDir(snapshotsDir)
			if err != nil {
				return fmt.Errorf("failed to read snapshots dir: %w", err)
			}
			var snapshotFiles []string
			for _, e := range entries {
				if !e.IsDir() && strings.HasPrefix(e.Name(), "amux-test-") && strings.HasSuffix(e.Name(), ".toml") {
					snapshotFiles = append(snapshotFiles, filepath.Join(snapshotsDir, e.Name()))
				}
			}
			if len(snapshotFiles) == 0 {
				return fmt.Errorf("no previous snapshots found for regression check")
			}

			sort.Strings(snapshotFiles)
			latestPath := snapshotFiles[len(snapshotFiles)-1]

			data, err := os.ReadFile(latestPath)
			if err != nil {
				return fmt.Errorf("failed to read latest snapshot: %w", err)
			}
			var oldSnapshot TestSnapshot
			if err := toml.Unmarshal(data, &oldSnapshot); err != nil {
				return fmt.Errorf("failed to parse latest snapshot: %w", err)
			}

			fmt.Printf("Comparing against baseline: %s (Phase %s)\n", latestPath, oldSnapshot.Phase)

			if currentSnapshot.Phase < oldSnapshot.Phase {
				return fmt.Errorf("REGRESSION: current phase %s < baseline phase %s", currentSnapshot.Phase, oldSnapshot.Phase)
			}
			fmt.Println("Regression check passed.")
			return nil
		}

		data, err := toml.Marshal(currentSnapshot)
		if err != nil {
			return err
		}

		if noSnapshotFlag {
			fmt.Println(string(data))
			return nil
		}

		filename := fmt.Sprintf("amux-test-%s.toml", currentSnapshot.Timestamp.Format("20060102-150405"))
		path := filepath.Join(snapshotsDir, filename)
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}

		fmt.Printf("Snapshot written to %s\n", path)
		return nil
	},
}
```


### Functions

#### findProjectRoot

```go
func findProjectRoot(startDir string) string
```

#### init

```go
func init()
```

#### main

```go
func main()
```


## type TestSnapshot

```go
type TestSnapshot struct {
	Timestamp time.Time `toml:"timestamp"`
	Version   string    `toml:"version"`
	Phase     string    `toml:"phase"`
}
```

TestSnapshot represents the verification snapshot.

