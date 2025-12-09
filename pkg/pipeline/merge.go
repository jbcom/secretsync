package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/jbcom/secretsync/pkg/client/vault"
	log "github.com/sirupsen/logrus"
)

// mergeTarget executes merge operations for a single target.
// Each merge is a distinct, idempotent operation: read from source(s), write to merge store.
func (p *Pipeline) mergeTarget(ctx context.Context, targetName string, dryRun bool) Result {
	start := time.Now()
	l := log.WithFields(log.Fields{
		"action": "mergeTarget",
		"target": targetName,
		"dryRun": dryRun,
	})

	target, ok := p.config.Targets[targetName]
	if !ok {
		return Result{
			Target:   targetName,
			Phase:    "merge",
			Success:  false,
			Error:    fmt.Errorf("target not found"),
			Duration: time.Since(start),
		}
	}

	// Determine merge path based on merge store type
	var mergePath string
	if p.config.MergeStore.Vault != nil {
		mergePath = fmt.Sprintf("%s/%s", p.config.MergeStore.Vault.Mount, targetName)
	} else if p.s3Store != nil {
		mergePath = p.s3Store.GetMergePath(targetName)
	} else {
		return Result{
			Target:   targetName,
			Phase:    "merge",
			Success:  false,
			Error:    fmt.Errorf("no merge store configured"),
			Duration: time.Since(start),
		}
	}
	l.WithField("mergePath", mergePath).Info("Starting merge")

	// Initialize Vault client for reading sources
	sourceClient := &vault.VaultClient{
		Address:   p.config.Vault.Address,
		Namespace: p.config.Vault.Namespace,
	}
	if err := sourceClient.Init(ctx); err != nil {
		return Result{
			Target:   targetName,
			Phase:    "merge",
			Success:  false,
			Error:    fmt.Errorf("failed to init source vault client: %w", err),
			Duration: time.Since(start),
		}
	}

	// Initialize Vault client for writing to merge store (if using Vault merge store)
	var mergeClient *vault.VaultClient
	if p.config.MergeStore.Vault != nil {
		mergeClient = &vault.VaultClient{
			Address:   p.config.Vault.Address,
			Namespace: p.config.Vault.Namespace,
		}
		if err := mergeClient.Init(ctx); err != nil {
			return Result{
				Target:   targetName,
				Phase:    "merge",
				Success:  false,
				Error:    fmt.Errorf("failed to init merge vault client: %w", err),
				Duration: time.Since(start),
			}
		}
	}

	var sourcePaths []string
	var failedImports []string
	var lastErr error
	successCount := 0

	// Each import is a distinct source→target operation
	for _, importName := range target.Imports {
		sourcePath := p.config.GetSourcePath(importName)
		sourcePaths = append(sourcePaths, sourcePath)

		l.WithFields(log.Fields{
			"import":     importName,
			"sourcePath": sourcePath,
		}).Debug("Processing import")

		// Read secrets from source
		secrets, err := sourceClient.ListSecrets(ctx, sourcePath)
		if err != nil {
			l.WithError(err).WithField("import", importName).Error("Failed to list secrets from source")
			failedImports = append(failedImports, importName)
			lastErr = err
			continue
		}

		if dryRun {
			l.WithFields(log.Fields{
				"import":       importName,
				"secretsCount": len(secrets),
			}).Info("[DRY-RUN] Would merge secrets")
			successCount++
			continue
		}

		// Write each secret to merge store
		for _, secretPath := range secrets {
			// Read secret data
			secretData, err := sourceClient.GetKVSecretOnce(ctx, secretPath)
			if err != nil {
				l.WithError(err).WithField("secret", secretPath).Warn("Failed to read secret, skipping")
				continue
			}

			// Write to merge store (Vault or S3)
			destPath := fmt.Sprintf("%s/%s", mergePath, secretPath)
			
			if mergeClient != nil {
				// Vault merge store
				if _, err := mergeClient.WriteSecretOnce(ctx, destPath, secretData, nil); err != nil {
					l.WithError(err).WithField("dest", destPath).Error("Failed to write to merge store")
					lastErr = err
					continue
				}
			} else if p.s3Store != nil {
				// S3 merge store
				if err := p.s3Store.WriteSecret(ctx, targetName, secretPath, secretData); err != nil {
					l.WithError(err).WithField("dest", destPath).Error("Failed to write to S3 merge store")
					lastErr = err
					continue
				}
			}
		}

		successCount++
	}

	success := lastErr == nil
	l.WithFields(log.Fields{
		"duration":      time.Since(start),
		"success":       success,
		"failedImports": failedImports,
	}).Info("Merge completed")

	result := Result{
		Target:    targetName,
		Phase:     "merge",
		Operation: string(OperationMerge),
		Success:   success,
		Error:     lastErr,
		Duration:  time.Since(start),
		Details: ResultDetails{
			SecretsProcessed: successCount,
			SourcePaths:      sourcePaths,
			DestinationPath:  mergePath,
			FailedImports:    failedImports,
		},
	}

	// Compute diff if tracking is enabled
	if p.pipelineDiff != nil {
		targetDiff, err := p.computeMergeDiff(ctx, targetName, sourcePaths)
		if err != nil {
			l.WithError(err).Debug("Failed to compute merge diff")
		} else {
			result.Diff = targetDiff
			p.addTargetDiff(*targetDiff)
		}
	}

	return result
}
