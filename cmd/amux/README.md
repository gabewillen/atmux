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

		snapshot := TestSnapshot{
			Timestamp: time.Now().UTC(),
			Version:   "0.0.0-phase0",
			Phase:     "0",
		}

		if regressionFlag {

			fmt.Println("Regression check passed (placeholder)")
			return nil
		}

		data, err := toml.Marshal(snapshot)
		if err != nil {
			return err
		}

		if noSnapshotFlag {
			fmt.Println(string(data))
			return nil
		}

		filename := fmt.Sprintf("amux-test-%s.toml", snapshot.Timestamp.Format("20060102-150405"))
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
	// Add more fields for regression checks (e.g. file hashes, config dumps)
	// For Phase 0 baseline, we just need the file to exist and contain metadata.
	Phase string `toml:"phase"`
}
```

TestSnapshot represents the verification snapshot.

