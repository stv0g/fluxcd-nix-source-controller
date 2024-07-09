// SPDX-FileCopyrightText: 2020, 2021 The Flux authors
// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	sourcev1 "github.com/fluxcd/source-controller/api/v1"
)

// GitRepositoryRevisionChangePredicate triggers an update event
// when a GitRepository revision changes.
type GitRepositoryRevisionChangePredicate struct {
	predicate.Funcs
}

func (GitRepositoryRevisionChangePredicate) Create(e event.CreateEvent) bool {
	src, ok := e.Object.(sourcev1.Source)

	if !ok || src.GetArtifact() == nil {
		return false
	}

	return true
}

func (GitRepositoryRevisionChangePredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}

	oldSource, ok := e.ObjectOld.(sourcev1.Source)
	if !ok {
		return false
	}

	newSource, ok := e.ObjectNew.(sourcev1.Source)
	if !ok {
		return false
	}

	if oldSource.GetArtifact() == nil && newSource.GetArtifact() != nil {
		return true
	}

	if oldSource.GetArtifact() != nil && newSource.GetArtifact() != nil &&
		oldSource.GetArtifact().Revision != newSource.GetArtifact().Revision {
		return true
	}

	return false
}
