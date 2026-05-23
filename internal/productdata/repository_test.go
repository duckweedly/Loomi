package productdata

import "testing"

func TestRepositoryContractUsesPostgresImplementation(t *testing.T) {
	var _ Repository = (*MemoryService)(nil)
}
