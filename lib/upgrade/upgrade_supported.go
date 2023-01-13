// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build !noupgrade
// +build !noupgrade

package upgrade

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const DisabledByCompilation = false

// FetchLatestReleases returns the latest releases. The "current" parameter
// is used for setting the User-Agent only.
func FetchLatestReleases(releasesURL, current string) []Release {
	return make([]Release, 0)
}

type SortByRelease []Release

func (s SortByRelease) Len() int {
	return len(s)
}
func (s SortByRelease) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortByRelease) Less(i, j int) bool {
	return CompareVersions(s[i].Tag, s[j].Tag) > 0
}

func LatestRelease(releasesURL, current string, upgradeToPreReleases bool) (Release, error) {
	rels := FetchLatestReleases(releasesURL, current)
	return SelectLatestRelease(rels, current, upgradeToPreReleases)
}

func SelectLatestRelease(rels []Release, current string, upgradeToPreReleases bool) (Release, error) {
	if len(rels) == 0 {
		return Release{}, ErrNoVersionToSelect
	}

	// Sort the releases, lowest version number first
	sort.Sort(sort.Reverse(SortByRelease(rels)))

	var selected Release
	for _, rel := range rels {
		if CompareVersions(rel.Tag, current) == MajorNewer {
			// We've found a new major version. That's fine, but if we've
			// already found a minor upgrade that is acceptable we should go
			// with that one first and then revisit in the future.
			if selected.Tag != "" && CompareVersions(selected.Tag, current) == Newer {
				return selected, nil
			}
		}

		if rel.Prerelease && !upgradeToPreReleases {
			l.Debugln("skipping pre-release", rel.Tag)
			continue
		}

		expectedReleases := releaseNames(rel.Tag)
	nextAsset:
		for _, asset := range rel.Assets {
			assetName := path.Base(asset.Name)
			// Check for the architecture
			for _, expRel := range expectedReleases {
				if strings.HasPrefix(assetName, expRel) {
					l.Debugln("selected", rel.Tag)
					selected = rel
					break nextAsset
				}
			}
		}
	}

	if selected.Tag == "" {
		return Release{}, ErrNoReleaseDownload
	}

	return selected, nil
}

// Upgrade to the given release, saving the previous binary with a ".old" extension.
func upgradeTo(binary string, rel Release) error {
	expectedReleases := releaseNames(rel.Tag)
	for _, asset := range rel.Assets {
		assetName := path.Base(asset.Name)
		l.Debugln("considering release", assetName)

		for _, expRel := range expectedReleases {
			if strings.HasPrefix(assetName, expRel) {
				return upgradeToURL(assetName, binary, asset.URL)
			}
		}
	}

	return ErrNoReleaseDownload
}

// Upgrade to the given release, saving the previous binary with a ".old" extension.
func upgradeToURL(archiveName, binary string, url string) error {
	fname, err := readRelease(archiveName, filepath.Dir(binary), url)
	if err != nil {
		return err
	}
	defer os.Remove(fname)

	old := binary + ".old"
	os.Remove(old)
	err = os.Rename(binary, old)
	if err != nil {
		return err
	}
	if err := os.Rename(fname, binary); err != nil {
		os.Rename(old, binary)
		return err
	}
	return nil
}

func readRelease(archiveName, dir, url string) (string, error) {
	return "", errors.New("autoupgrade disabled")
}
