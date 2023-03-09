/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package version provide api for versioning information.
package version

var (
	version   = "version-not-set"
	gitRev    = "git-rev-not-set"    //nolint:gochecknoglobals
	buildTime = "build-time-not-set" //nolint:gochecknoglobals
)

// GetVersion returns sdk version.
func GetVersion() string {
	return version
}

// GetGitRevision returns sdk git revision.
func GetGitRevision() string {
	return gitRev
}

// GetBuildTime returns sdk build time.
func GetBuildTime() string {
	return buildTime
}
