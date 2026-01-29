package inference

/*
#cgo CFLAGS: -I${SRCDIR}/../../third_party/liquidgen/src/inference -I${SRCDIR}/../../third_party/liquidgen/src/ggml -I${SRCDIR}/../../third_party/liquidgen/src/loader -I${SRCDIR}/../../third_party/liquidgen/src
#cgo LDFLAGS: -L${SRCDIR}/../../third_party/liquidgen/build/lib -lliquid_core -lliquid_vision -llfm_ggml -lggml-cpu -llfm_ggml_base -lm -lstdc++ -ldl

#include "lfm_inference.h"
#include <stdlib.h>

// Helper to get default params which are inline/macro sometimes or return structs
static struct liquid_model_params get_default_mparams() {
    return liquid_model_default_params();
}

static struct liquid_context_params get_default_cparams() {
    return liquid_context_default_params();
}
*/
import "C"
import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"unsafe"
)

var (
	backendOnce sync.Once
)

func init() {
	// Spec §4.2.10: build and runtime logs MUST include the liquidgen module version or commit identifier
	fmt.Printf("liquidgen module version: %s\n", "v0.0.1-dev-integrated")
}

type LiquidgenImpl struct {
	mu     sync.Mutex
	models map[string]*C.struct_liquid_model
}

func NewLiquidgenImpl() *LiquidgenImpl {
	backendOnce.Do(func() {
		C.liquid_backend_init()
	})
	return &LiquidgenImpl{
		models: make(map[string]*C.struct_liquid_model),
	}
}

func (e *LiquidgenImpl) Generate(ctx context.Context, req LiquidgenRequest) (LiquidgenStream, error) {
	e.mu.Lock()
	model, ok := e.models[req.Model]
	if !ok {
		path, err := resolveModelPath(req.Model)
		if err != nil {
			e.mu.Unlock()
			return nil, err
		}
		cPath := C.CString(path)
		defer C.free(unsafe.Pointer(cPath))
		
		mparams := C.get_default_mparams()
		model = C.liquid_model_load_from_file(cPath, mparams)
		if model == nil {
			e.mu.Unlock()
			return nil, fmt.Errorf("failed to load model from %s", path)
		}
		e.models[req.Model] = model
	}
	e.mu.Unlock()

	// Init context
	cparams := C.get_default_cparams()
	if req.MaxTokens > 0 {
		cparams.n_ctx = C.uint32_t(req.MaxTokens + 512) // Rough buffer
	}
	
	lctx := C.liquid_init_from_model(model, cparams)
	if lctx == nil {
		return nil, fmt.Errorf("failed to init liquid context")
	}

	vocab := C.liquid_model_get_vocab(model)
	
	// Tokenize prompt
	cPrompt := C.CString(req.Prompt)
	defer C.free(unsafe.Pointer(cPrompt))
	
	// Pre-allocate token buffer
	nTokensMax := 4096
	tokens := make([]C.liquid_token, nTokensMax)
	nTokens := C.liquid_tokenize(vocab, cPrompt, C.int32_t(len(req.Prompt)), (*C.liquid_token)(&tokens[0]), C.int32_t(nTokensMax), true, true)
	if nTokens < 0 {
		C.liquid_free(lctx)
		return nil, fmt.Errorf("tokenization failed (need %d tokens)", -nTokens)
	}

	// Create sampler
	sparams := C.liquid_sampler_chain_default_params()
	sampler := C.liquid_sampler_chain_init(sparams)
	C.liquid_sampler_chain_add(sampler, C.liquid_sampler_init_temp(C.float(req.Temperature)))
	C.liquid_sampler_chain_add(sampler, C.liquid_sampler_init_dist(C.uint32_t(C.LIQUID_DEFAULT_SEED)))

	stream := &liquidgenStream{
		ctx:     ctx,
		lctx:    lctx,
		model:   model,
		vocab:   vocab,
		sampler: sampler,
		max:     req.MaxTokens,
	}

	// Ingest initial tokens
	batch := C.liquid_batch_get_one((*C.liquid_token)(&tokens[0]), nTokens)
	if res := C.liquid_decode(lctx, batch); res != 0 {
		stream.Close()
		return nil, fmt.Errorf("initial decode failed: %d", res)
	}
	stream.nPast = int(nTokens)

	return stream, nil
}

type liquidgenStream struct {
	ctx     context.Context
	lctx    *C.struct_liquid_context
	model   *C.struct_liquid_model
	vocab   *C.struct_liquid_vocab
	sampler *C.struct_liquid_sampler
	nPast   int
	max     int
	count   int
	closed  bool
}

func (s *liquidgenStream) Next() (string, error) {
	if s.closed {
		return "", io.EOF
	}
	if s.max > 0 && s.count >= s.max {
		return "", io.EOF
	}

	// Sample
	token := C.liquid_sampler_sample(s.sampler, s.lctx, -1)
	if C.liquid_vocab_is_eog(s.vocab, token) {
		return "", io.EOF
	}

	// Detokenize
	buf := make([]byte, 128)
	n := C.liquid_token_to_piece(s.vocab, token, (*C.char)(unsafe.Pointer(&buf[0])), C.int32_t(len(buf)), 0, true)
	if n < 0 {
		return "", fmt.Errorf("detokenization failed")
	}
	piece := string(buf[:n])

	// Decode next
	batch := C.liquid_batch_get_one(&token, 1)
	if res := C.liquid_decode(s.lctx, batch); res != 0 {
		return piece, fmt.Errorf("decode failed: %d", res)
	}

	s.nPast++
	s.count++
	return piece, nil
}

func (s *liquidgenStream) Close() error {
	if !s.closed {
		C.liquid_sampler_free(s.sampler)
		C.liquid_free(s.lctx)
		s.closed = true
	}
	return nil
}

func resolveModelPath(modelID string) (string, error) {
	mapping := map[string]string{
		"lfm2.5-thinking": "LFM2.5-1.2B-Thinking-GGUF/LFM2.5-1.2B-Thinking-Q4_K_M.gguf",
		"lfm2.5-VL":       "LFM2.5-VL-1.6B-GGUF/LFM2.5-VL-1.6B-Q4_K_M.gguf",
	}
	
	relPath, ok := mapping[modelID]
	if !ok {
		return "", fmt.Errorf("unknown model ID: %s", modelID)
	}

	home := os.Getenv("HOME")
	searchPaths := []string{
		"models",
		filepath.Join(home, ".amux", "models"),
		"/shared/gemini-cli-auto/models", // Absolute path in this env
	}
	
	for _, p := range searchPaths {
		fullPath := filepath.Join(p, relPath)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}
	
	return "", fmt.Errorf("model %s not found", modelID)
}
