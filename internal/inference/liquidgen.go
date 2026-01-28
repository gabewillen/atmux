package inference

import (
	"context"
	"fmt"
)

type LiquidgenImpl struct{}

func (e *LiquidgenImpl) Generate(ctx context.Context, req LiquidgenRequest) (LiquidgenStream, error) {
	return nil, fmt.Errorf("liquidgen integration not yet implemented")
}
