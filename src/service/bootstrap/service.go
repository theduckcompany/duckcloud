package bootstrap

import "context"

type BootstrapService struct {
}

// NewBootstrapService instantiates a new [Service].
func NewBootstrapService() *BootstrapService {
	return &BootstrapService{}
}

func (s *BootstrapService) Bootstrap(ctx context.Context) error {
	return nil
}
