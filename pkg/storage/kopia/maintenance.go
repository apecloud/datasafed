package kopia

import (
	"context"
	"fmt"

	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/repo/maintenance"
	"github.com/kopia/kopia/snapshot/snapshotmaintenance"

	"github.com/apecloud/datasafed/pkg/storage"
)

type unwrap interface {
	Unwrap() storage.Storage
}

func asKopiaStorage(st storage.Storage) (*kopiaStorage, bool) {
	for {
		if ks, ok := st.(*kopiaStorage); ok {
			return ks, true
		}
		if u, ok := st.(unwrap); ok {
			st = u.Unwrap()
			continue
		}
		return nil, false
	}
}

func RunMaintenance(ctx context.Context, st storage.Storage, safety string) error {
	ks, ok := asKopiaStorage(st)
	if !ok {
		return fmt.Errorf("requires *kopiaStorage, got %T", st)
	}

	directRep, ok := ks.rep.(repo.DirectRepository)
	if !ok {
		return fmt.Errorf("requires repo.DirectRepository, got %T", ks.rep)
	}

	return repo.DirectWriteSession(ctx, directRep, repo.WriteSessionOptions{
		Purpose: "datasafed:maintenance",
	}, func(ctx context.Context, dw repo.DirectRepositoryWriter) error {
		mode := maintenance.ModeQuick
		_, supportsEpochManager, err := dw.ContentManager().EpochManager()
		if err != nil {
			return fmt.Errorf("EpochManager error: %w", err)
		}
		if supportsEpochManager {
			mode = maintenance.ModeFull
		}
		safetyParams := maintenance.SafetyFull
		if safety == "none" {
			safetyParams = maintenance.SafetyNone
		} else {
			safety = "full" // default full
		}
		log(ctx).Infof("[KOPIA] maintenance mode: %s, safety: %s", mode, safety)
		// set force to true to ignore the ownership checking
		force := true
		return snapshotmaintenance.Run(ctx, dw, mode, force, safetyParams)
	})
}
