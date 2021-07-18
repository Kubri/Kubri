package local_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/abemedia/appcast/source"
	"github.com/abemedia/appcast/source/local"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestFile(t *testing.T) {
	dir := t.TempDir()

	want := []*source.Release{
		{
			Name:    "v0.0.0",
			Version: "v0.0.0",
			Assets: []*source.Asset{
				{
					Name: "test.dmg",
					URL:  "file://" + filepath.Join(dir, "test.dmg"),
					Size: 5,
				},
				{
					Name: "test_32-bit.msi",
					URL:  "file://" + filepath.Join(dir, "test_32-bit.msi"),
					Size: 5,
				},
				{
					Name: "test_64-bit.msi",
					URL:  "file://" + filepath.Join(dir, "test_64-bit.msi"),
					Size: 5,
				},
			},
		},
	}

	for _, release := range want {
		for _, asset := range release.Assets {
			err := os.WriteFile(filepath.Join(dir, asset.Name), []byte("test\n"), os.ModePerm)
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	r, err := local.New(source.Config{Repo: dir})
	if err != nil {
		t.Fatal(err)
	}

	opt := cmpopts.IgnoreFields(source.Release{}, "Date")

	t.Run("ListReleases", func(t *testing.T) {
		got, err := r.ListReleases(nil)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(want, got, opt); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("GetRelease", func(t *testing.T) {
		got, err := r.GetRelease(want[0].Version)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(want[0], got, opt); diff != "" {
			t.Error(diff)
		}
	})

	asset := []byte("foo")

	t.Run("UploadAsset", func(t *testing.T) {
		err := r.UploadAsset(want[0].Version, "test.txt", asset)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("DownloadAsset", func(t *testing.T) {
		b, err := r.DownloadAsset(want[0].Version, "test.txt")
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(asset, b) {
			t.Error("should be equal")
		}
	})
}
