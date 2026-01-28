package inference

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ErrUnknownModel is returned when a model ID is not mapped.
var ErrUnknownModel = errors.New("unknown model id")

// LiquidgenEngine implements Engine using the bundled liquidgen runtime.
type LiquidgenEngine struct {
	root      string
	version   string
	models    map[string]string
	logger    *log.Logger
}

// NewLiquidgenEngine loads the liquidgen engine from the provided root.
func NewLiquidgenEngine(root string, logger *log.Logger) (*LiquidgenEngine, error) {
	if logger == nil {
		logger = log.New(os.Stderr, "amux-liquidgen ", log.LstdFlags)
	}
	version, err := readLiquidgenVersion(root)
	if err != nil {
		return nil, fmt.Errorf("liquidgen version: %w", err)
	}
	engine := &LiquidgenEngine{
		root:    root,
		version: version,
		models:  make(map[string]string),
		logger:  logger,
	}
	engine.logger.Printf("liquidgen version=%s", version)
	return engine, nil
}

// Version returns the liquidgen version or commit identifier.
func (l *LiquidgenEngine) Version() string {
	return l.version
}

// RegisterModel maps a logical model ID to an artifact path.
func (l *LiquidgenEngine) RegisterModel(id string, artifactPath string) {
	l.models[id] = artifactPath
	l.logger.Printf("liquidgen model id=%s path=%s", id, artifactPath)
}

// Models returns the registered model mappings.
func (l *LiquidgenEngine) Models(ctx context.Context) ([]ModelInfo, error) {
	models := make([]ModelInfo, 0, len(l.models))
	for id, path := range l.models {
		models = append(models, ModelInfo{ID: id, ArtifactPath: path})
	}
	tracer := otel.Tracer("amux.inference")
	_, span := tracer.Start(ctx, "inference.models")
	for _, model := range models {
		span.AddEvent("model.mapping", trace.WithAttributes(
			attribute.String("model.id", model.ID),
			attribute.String("model.path", model.ArtifactPath),
		))
	}
	span.End()
	return models, nil
}

// Infer returns an error for unknown models until liquidgen runtime is wired.
func (l *LiquidgenEngine) Infer(ctx context.Context, req Request) (Response, error) {
	_, ok := l.models[req.ModelID]
	if !ok {
		return Response{}, fmt.Errorf("infer: %w: %s", ErrUnknownModel, req.ModelID)
	}
	return Response{}, fmt.Errorf("infer: liquidgen runtime not yet wired")
}

func readLiquidgenVersion(root string) (string, error) {
	path := filepath.Join(root, "CMakeLists.txt")
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(file)
	var version string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "project(liquidgen") && strings.Contains(line, "VERSION") {
			version = extractVersion(line)
			if version != "" {
				break
			}
		}
	}
	closeErr := file.Close()
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if closeErr != nil {
		return "", closeErr
	}
	if version == "" {
		return "", fmt.Errorf("version not found")
	}
	return version, nil
}

func extractVersion(line string) string {
	fields := strings.Fields(line)
	for i, field := range fields {
		if strings.EqualFold(field, "VERSION") && i+1 < len(fields) {
			return strings.Trim(fields[i+1], ")")
		}
	}
	return ""
}
