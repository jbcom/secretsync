package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jbcom/secretsync/pkg/client/aws"
	"github.com/jbcom/secretsync/pkg/client/vault"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// syncTarget executes sync operations for a single target.
// Each sync is a distinct, idempotent operation: read from merge store, write to AWS.
func (p *Pipeline) syncTarget(ctx context.Context, targetName string, dryRun bool) Result {
	start := time.Now()
	l := log.WithFields(log.Fields{
		"action": "syncTarget",
		"target": targetName,
		"dryRun": dryRun,
	})

	target, ok := p.config.Targets[targetName]
	if !ok {
		return Result{
			Target:   targetName,
			Phase:    "sync",
			Success:  false,
			Error:    fmt.Errorf("target not found"),
			Duration: time.Since(start),
		}
	}

	// Determine source (merge store) path
	var mergePath string
	if p.config.MergeStore.Vault != nil {
		mergePath = fmt.Sprintf("%s/%s", p.config.MergeStore.Vault.Mount, targetName)
	} else if p.s3Store != nil {
		mergePath = p.s3Store.GetMergePath(targetName)
	} else {
		return Result{
			Target:   targetName,
			Phase:    "sync",
			Success:  false,
			Error:    fmt.Errorf("no merge store configured"),
			Duration: time.Since(start),
		}
	}

	l.WithFields(log.Fields{
		"mergePath": mergePath,
		"accountId": target.AccountID,
	}).Info("Starting sync")

	// Initialize source client (merge store)
	var secrets []string
	var secretsData map[string]map[string]interface{}

	if p.config.MergeStore.Vault != nil {
		mergeClient := &vault.VaultClient{
			Address:   p.config.Vault.Address,
			Namespace: p.config.Vault.Namespace,
		}
		if err := mergeClient.Init(ctx); err != nil {
			return Result{
				Target:   targetName,
				Phase:    "sync",
				Success:  false,
				Error:    fmt.Errorf("failed to init merge vault client: %w", err),
				Duration: time.Since(start),
			}
		}

		var err error
		secrets, err = mergeClient.ListSecrets(ctx, mergePath)
		if err != nil {
			return Result{
				Target:   targetName,
				Phase:    "sync",
				Success:  false,
				Error:    fmt.Errorf("failed to list secrets from merge store: %w", err),
				Duration: time.Since(start),
			}
		}

		// Read all secret data
		secretsData = make(map[string]map[string]interface{})
		for _, secretPath := range secrets {
			data, err := mergeClient.GetKVSecretOnce(ctx, secretPath)
			if err != nil {
				l.WithError(err).WithField("secret", secretPath).Warn("Failed to read secret from merge store")
				continue
			}
			secretsData[secretPath] = data
		}
	} else if p.s3Store != nil {
		var err error
		secrets, err = p.s3Store.ListSecrets(ctx, targetName)
		if err != nil {
			return Result{
				Target:   targetName,
				Phase:    "sync",
				Success:  false,
				Error:    fmt.Errorf("failed to list secrets from S3 merge store: %w", err),
				Duration: time.Since(start),
			}
		}

		secretsData = make(map[string]map[string]interface{})
		for _, secretPath := range secrets {
			data, err := p.s3Store.ReadSecret(ctx, targetName, secretPath)
			if err != nil {
				l.WithError(err).WithField("secret", secretPath).Warn("Failed to read secret from S3")
				continue
			}
			secretsData[secretPath] = data
		}
	}

	l.WithField("secretsCount", len(secrets)).Debug("Retrieved secrets from merge store")

	if dryRun {
		l.WithField("secretsCount", len(secrets)).Info("[DRY-RUN] Would sync secrets to AWS")
		return Result{
			Target:    targetName,
			Phase:     "sync",
			Operation: string(OperationSync),
			Success:   true,
			Duration:  time.Since(start),
			Details: ResultDetails{
				SecretsProcessed: len(secrets),
				SourcePaths:      []string{mergePath},
				DestinationPath:  fmt.Sprintf("aws://%s", target.AccountID),
			},
		}
	}

	// Get role ARN and region for this target
	roleARN := p.getRoleARNForTarget(target)
	region := target.Region
	if region == "" {
		region = p.config.AWS.Region
	}

	// Initialize AWS client for target account
	awsClient, err := p.getAWSClientForTarget(ctx, target)
	if err != nil {
		return Result{
			Target:   targetName,
			Phase:    "sync",
			Success:  false,
			Error:    fmt.Errorf("failed to get AWS client for target: %w", err),
			Duration: time.Since(start),
		}
	}

	// Sync each secret to AWS
	var syncErrors []string
	successCount := 0

	for secretPath, data := range secretsData {
		// Determine AWS secret name
		awsSecretName := p.getAWSSecretName(targetName, secretPath)

		// Convert data to JSON bytes for AWS
		secretBytes, err := json.Marshal(data)
		if err != nil {
			l.WithError(err).WithField("secret", secretPath).Error("Failed to marshal secret data")
			syncErrors = append(syncErrors, secretPath)
			continue
		}

		// Create metadata for the write operation
		meta := metav1.ObjectMeta{
			Name:      awsSecretName,
			Namespace: targetName,
		}

		if _, err := awsClient.WriteSecret(ctx, meta, awsSecretName, secretBytes); err != nil {
			l.WithError(err).WithFields(log.Fields{
				"secret":    secretPath,
				"awsSecret": awsSecretName,
			}).Error("Failed to write secret to AWS")
			syncErrors = append(syncErrors, secretPath)
			continue
		}

		l.WithFields(log.Fields{
			"secret":    secretPath,
			"awsSecret": awsSecretName,
		}).Debug("Secret synced to AWS")
		successCount++
	}

	success := len(syncErrors) == 0
	var lastErr error
	if !success {
		lastErr = fmt.Errorf("failed to sync %d secrets: %v", len(syncErrors), syncErrors)
	}

	l.WithFields(log.Fields{
		"duration": time.Since(start),
		"success":  success,
		"synced":   successCount,
		"failed":   len(syncErrors),
	}).Info("Sync completed")

	result := Result{
		Target:    targetName,
		Phase:     "sync",
		Operation: string(OperationSync),
		Success:   success,
		Error:     lastErr,
		Duration:  time.Since(start),
		Details: ResultDetails{
			SecretsProcessed: successCount,
			SourcePaths:      []string{mergePath},
			DestinationPath:  fmt.Sprintf("aws://%s", target.AccountID),
			RoleARN:          roleARN,
		},
	}

	// Compute diff if tracking is enabled
	if p.pipelineDiff != nil {
		targetDiff, err := p.computeSyncDiff(ctx, targetName, roleARN, region)
		if err != nil {
			l.WithError(err).Debug("Failed to compute sync diff")
		} else {
			result.Diff = targetDiff
			p.addTargetDiff(*targetDiff)
		}
	}

	return result
}

// getAWSClientForTarget returns an AWS client configured for the target account.
// It handles cross-account role assumption via Control Tower or custom patterns.
func (p *Pipeline) getAWSClientForTarget(ctx context.Context, target Target) (*aws.AwsClient, error) {
	region := target.Region
	if region == "" {
		region = p.config.AWS.Region
	}

	client := &aws.AwsClient{
		Region: region,
	}

	// If we have an AWS execution context with role assumption
	roleArn := p.getRoleARNForTarget(target)
	if roleArn != "" {
		client.RoleArn = roleArn
	}

	if err := client.Init(ctx); err != nil {
		return nil, err
	}

	return client, nil
}

// getRoleARNForTarget returns the role ARN for assuming into the target account
func (p *Pipeline) getRoleARNForTarget(target Target) string {
	if target.AccountID == "" {
		return ""
	}

	// Use custom role pattern if provided
	if p.awsCtx != nil && p.config.AWS.ExecutionContext.CustomRolePattern != "" {
		return fmt.Sprintf(p.config.AWS.ExecutionContext.CustomRolePattern, target.AccountID)
	}

	// Use Control Tower execution role if enabled
	if p.config.AWS.ControlTower.Enabled {
		roleName := p.config.AWS.ControlTower.ExecutionRole.Name
		if roleName == "" {
			roleName = "AWSControlTowerExecution"
		}
		return fmt.Sprintf("arn:aws:iam::%s:role/%s", target.AccountID, roleName)
	}

	return ""
}

// getAWSSecretName determines the AWS Secrets Manager secret name for a given path.
func (p *Pipeline) getAWSSecretName(targetName, secretPath string) string {
	// Default: use the secret path as-is
	// Could be customized via target config or naming patterns
	return secretPath
}
