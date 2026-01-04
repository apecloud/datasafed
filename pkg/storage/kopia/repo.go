package kopia

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/repo/blob"
	"github.com/kopia/kopia/repo/ecc"
	"github.com/kopia/kopia/repo/encryption"
	"github.com/kopia/kopia/repo/format"
	"github.com/kopia/kopia/repo/hashing"
	"github.com/kopia/kopia/repo/maintenance"
	"github.com/kopia/kopia/snapshot/policy"
)

const (
	formatVersion        = format.FormatVersion3
	blockHashAlgorithm   = hashing.DefaultAlgorithm
	encryptionAlgorithm  = encryption.DefaultAlgorithm
	eccAlgorithm         = ecc.DefaultAlgorithm
	splitterAlgorithm    = "DYNAMIC-512K-BUZHASH" // the granularity of the default splitter is too rough
	compressionAlgorithm = "zstd"
	defaultPassword      = "d@ta$aFed"
)

func getInitedRepository(ctx context.Context, rootPath, password, cacheID string) (repo.Repository, error) {
	opts := &blobOptions{RootPath: rootPath}
	configFile, err := writeTempConfigFile(opts, cacheID)
	if err != nil {
		return nil, err
	}
	if password == "" {
		password = defaultPassword
	}
	rep, err := repo.Open(ctx, configFile, password, &repo.Options{})
	if errors.Is(err, blob.ErrBlobNotFound) {
		st, err := newBlobStorage(ctx, opts, true)
		if err != nil {
			return nil, err
		}
		return initRepository(ctx, st, configFile, password)
	}
	if err != nil {
		return nil, err
	}
	return rep, nil
}

func writeTempConfigFile(opt *blobOptions, cacheID string) (string, error) {
	// create temporary config file
	f, err := os.CreateTemp("", "config*.json")
	if err != nil {
		return "", fmt.Errorf("create temporary config file failed, err: %w", err)
	}
	defer f.Close()
	// Doc: https://kopia.io/docs/reference/command-line/#configuration-file
	conf := map[string]interface{}{
		"storage": map[string]interface{}{
			"type":   storageType,
			"config": opt,
		},
	}
	if cacheID != "" {
		cacheDir := filepath.Join(os.TempDir(), "kopiacache", cacheID)
		conf["caching"] = map[string]interface{}{
			"cacheDirectory":       cacheDir,
			"maxCacheSize":         128 * 1024 * 1024,
			"maxMetadataCacheSize": 32 * 1024 * 1024,
			"maxListCacheDuration": 60,
		}
	}

	err = json.NewEncoder(f).Encode(conf)
	if err != nil {
		return "", fmt.Errorf("encode config file failed, err: %w", err)
	}
	return f.Name(), nil
}

func ensureEmpty(ctx context.Context, s blob.Storage) error {
	hasDataError := errors.New("has data")

	err := s.ListBlobs(ctx, "", func(cb blob.Metadata) error {
		return hasDataError
	})

	if errors.Is(err, hasDataError) {
		return errors.New("found existing data in storage location")
	}

	if err != nil {
		return fmt.Errorf("listing blobs error: %s", err)
	}
	return nil
}

func initRepository(ctx context.Context, st blob.Storage, configFile string, password string) (repo.Repository, error) {
	err := ensureEmpty(ctx, st)
	if err != nil {
		return nil, fmt.Errorf("initRepostory: ensureEmpty failed, %w", err)
	}

	options := &repo.NewRepositoryOptions{
		BlockFormat: format.ContentFormat{
			MutableParameters: format.MutableParameters{
				Version: formatVersion,
			},
			Hash:               blockHashAlgorithm,
			Encryption:         encryptionAlgorithm,
			ECC:                eccAlgorithm,
			ECCOverheadPercent: 0, // disable ECC
		},

		ObjectFormat: format.ObjectFormat{
			Splitter: splitterAlgorithm,
		},

		// no retention
		RetentionMode:   "",
		RetentionPeriod: 0,
	}

	if err := repo.Initialize(ctx, st, options, password); err != nil {
		return nil, fmt.Errorf("initRepostory: cannot initialize repository, %w", err)
	}

	rep, err := repo.Open(ctx, configFile, password, &repo.Options{})
	if err != nil {
		return nil, fmt.Errorf("unable to open repository, %w", err)
	}

	if err := populateRepository(ctx, rep); err != nil {
		return nil, fmt.Errorf("error populating repository, %w", err)
	}
	return rep, nil
}

func populateRepository(ctx context.Context, rep repo.Repository) error {
	err := repo.WriteSession(ctx, rep, repo.WriteSessionOptions{
		Purpose: "populate repository",
	}, func(ctx context.Context, w repo.RepositoryWriter) error {
		myPolicy := *policy.DefaultPolicy
		myPolicy.CompressionPolicy = policy.CompressionPolicy{
			CompressorName: compressionAlgorithm,
		}
		if err := policy.SetPolicy(ctx, w, policy.GlobalPolicySourceInfo, &myPolicy); err != nil {
			return fmt.Errorf("unable to set global policy, %w", err)
		}

		if err := setDefaultMaintenanceParameters(ctx, w); err != nil {
			return fmt.Errorf("unable to set maintenance parameters, %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to write session, %w", err)
	}
	return nil
}

func setDefaultMaintenanceParameters(ctx context.Context, rep repo.RepositoryWriter) error {
	p := maintenance.DefaultParams()
	p.Owner = rep.ClientOptions().UsernameAtHost()

	if dw, ok := rep.(repo.DirectRepositoryWriter); ok {
		_, ok, err := dw.ContentReader().EpochManager(ctx)
		if err != nil {
			return fmt.Errorf("epoch manager, %w", err)
		}

		if ok {
			// disable quick maintenance cycle
			p.QuickCycle.Enabled = false
		}
	}

	if err := maintenance.SetParams(ctx, rep, &p); err != nil {
		return fmt.Errorf("unable to set maintenance params, %w", err)
	}

	return nil
}
