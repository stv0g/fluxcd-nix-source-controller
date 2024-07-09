// SPDX-FileCopyrightText: 2020, 2021 The Flux authors
// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"os"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluxcd/pkg/http/fetch"
	"github.com/fluxcd/pkg/tar"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	tarx "github.com/stv0g/flux-nix-controller/internal/tar"
	"github.com/stv0g/flux-nix-controller/pkg/nix"
)

// GitRepositoryWatcher watches GitRepository objects for revision changes
type GitRepositoryWatcher struct {
	client.Client
	artifactFetcher *fetch.ArchiveFetcher
	HttpRetry       int
}

func (r *GitRepositoryWatcher) SetupWithManager(mgr ctrl.Manager) error {
	r.artifactFetcher = fetch.New(
		fetch.WithRetries(r.HttpRetry),
		fetch.WithMaxDownloadSize(tar.UnlimitedUntarSize),
		fetch.WithUntar(tar.WithMaxUntarSize(tar.UnlimitedUntarSize)),
		fetch.WithHostnameOverwrite(os.Getenv("SOURCE_CONTROLLER_LOCALHOST")),
		fetch.WithLogger(nil),
	)

	return ctrl.NewControllerManagedBy(mgr).
		For(&sourcev1.GitRepository{}, builder.WithPredicates(GitRepositoryRevisionChangePredicate{})).
		Complete(r)
}

// +kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=gitrepositories,verbs=get;list;watch
// +kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=gitrepositories/status,verbs=get

func (r *GitRepositoryWatcher) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Get source object
	var repository sourcev1.GitRepository
	if err := r.Get(ctx, req.NamespacedName, &repository); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	artifact := repository.Status.Artifact
	log.Info("New revision detected", "revision", artifact.Revision)

	// Create tmp dir
	tmpDir, err := os.MkdirTemp("", repository.Name)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create temp dir, error: %w", err)
	}

	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			log.Error(err, "unable to remove temp dir")
		}
	}(tmpDir)

	// Download and extract artifact
	if err := r.artifactFetcher.Fetch(artifact.URL, artifact.Digest, tmpDir); err != nil {
		log.Error(err, "unable to fetch artifact")
		return ctrl.Result{}, err
	}

	results, err := nix.Build(nil, nil, []string{tmpDir})

	if len(results) != 1 {
		return ctrl.Result{}, fmt.Errorf("invalid number of results")
	}

	output, ok := results[0].Outputs["out"]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("missing output 'out'")
	}

	archive := &bytes.Buffer{}

	if err := tarx.Compress(os.DirFS(output), archive); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to compress: %w", err)
	}

	minioClient, err := minio.New("myendpoint", &minio.Options{
		Creds:  credentials.NewStaticV4("id", "secret", ""),
		Secure: true,
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create Minio client: %w", err)
	}

	ui, err := minioClient.PutObject(context.Background(), "bucket", "object.tar.gz", archive, int64(archive.Len()), minio.PutObjectOptions{})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to upload archive to bucket: %w", err)
	}

	return ctrl.Result{}, nil
}
