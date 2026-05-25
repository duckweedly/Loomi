package runtime

import (
	"context"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type MemoryProvider interface {
	SearchMemory(context.Context, productdata.MemorySearchInput) (productdata.MemorySearchOutput, error)
	BuildSnapshot(context.Context, productdata.Run, productdata.Thread) (productdata.MemorySnapshot, error)
	CreateWriteProposal(context.Context, productdata.ProposeMemoryWriteInput) (productdata.MemoryWriteProposal, error)
	ApproveWrite(context.Context, string, productdata.MemoryWriteDecisionInput) (productdata.MemoryWriteDecision, error)
	DenyWrite(context.Context, string, productdata.MemoryWriteDecisionInput) (productdata.MemoryWriteDecision, error)
	DeleteMemory(context.Context, string, productdata.DeleteMemoryEntryInput) (productdata.MemoryTombstone, error)
}

type ProductMemoryProvider struct {
	Service productdata.Service
	Ident   identity.LocalIdentity
}

func (p ProductMemoryProvider) ident() identity.LocalIdentity {
	if p.Ident.UserID == "" {
		return identity.LocalDevIdentity()
	}
	return p.Ident
}

func (p ProductMemoryProvider) SearchMemory(ctx context.Context, input productdata.MemorySearchInput) (productdata.MemorySearchOutput, error) {
	return p.Service.SearchMemory(ctx, p.ident(), input)
}

func (p ProductMemoryProvider) BuildSnapshot(ctx context.Context, run productdata.Run, thread productdata.Thread) (productdata.MemorySnapshot, error) {
	output, err := p.SearchMemory(ctx, productdata.MemorySearchInput{ScopeType: productdata.MemoryScopeThread, ScopeID: thread.ID, Limit: 5, Purpose: "run_context"})
	if err != nil {
		return productdata.MemorySnapshot{RunID: run.ID, ThreadID: thread.ID, Limit: 5, LoadStatus: "unavailable"}, err
	}
	status := "loaded"
	if len(output.Items) == 0 {
		status = "empty"
	}
	return productdata.MemorySnapshot{RunID: run.ID, ThreadID: thread.ID, Entries: output.Items, Limit: 5, TotalCandidates: len(output.Items), LoadStatus: status, RedactionApplied: true}, nil
}

func (p ProductMemoryProvider) CreateWriteProposal(ctx context.Context, input productdata.ProposeMemoryWriteInput) (productdata.MemoryWriteProposal, error) {
	return p.Service.ProposeMemoryWrite(ctx, p.ident(), input)
}

func (p ProductMemoryProvider) ApproveWrite(ctx context.Context, proposalID string, input productdata.MemoryWriteDecisionInput) (productdata.MemoryWriteDecision, error) {
	return p.Service.ApproveMemoryWrite(ctx, p.ident(), proposalID, input)
}

func (p ProductMemoryProvider) DenyWrite(ctx context.Context, proposalID string, input productdata.MemoryWriteDecisionInput) (productdata.MemoryWriteDecision, error) {
	return p.Service.DenyMemoryWrite(ctx, p.ident(), proposalID, input)
}

func (p ProductMemoryProvider) DeleteMemory(ctx context.Context, entryID string, input productdata.DeleteMemoryEntryInput) (productdata.MemoryTombstone, error) {
	return p.Service.DeleteMemoryEntry(ctx, p.ident(), entryID, input)
}
