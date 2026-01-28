# Models Directory

This directory contains ONNX models and runtime libraries for local inference.

## Structure
- `onnxruntime/` - ONNX Runtime shared libraries by platform
- `embeddings/` - Embedding models (e.g., all-MiniLM-L6-v2)
- `liquidgen/` - liquidgen model artifacts

## Implementation Status
Model assets will be packaged in Phase 6 for semantic subscriptions.

## Required Models
Per spec §4.2.10:
- `lfm2.5-thinking` (text-only reasoning)
- `lfm2.5-VL` (vision-language)

Default embedding model: all-MiniLM-L6-v2