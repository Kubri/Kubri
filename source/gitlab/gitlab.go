package gitlab

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"path"
	"strconv"

	"github.com/abemedia/appcast/source"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/xanzy/go-gitlab"
)

type gitlabSource struct {
	client *gitlab.Client
	repo   string
}

func New(c source.Config) (*source.Source, error) {
	s := new(gitlabSource)

	client, err := gitlab.NewClient(c.Token)
	if err != nil {
		return nil, err
	}

	s.client = client
	s.repo = c.Repo

	return &source.Source{Provider: s}, nil
}

func (s *gitlabSource) ListReleases(ctx context.Context) ([]*source.Release, error) {
	releases, _, err := s.client.Releases.ListReleases(s.repo, nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	r := make([]*source.Release, 0, len(releases))
	for _, release := range releases {
		r = append(r, s.parseRelease(ctx, release))
	}

	return r, nil
}

func (s *gitlabSource) GetRelease(ctx context.Context, version string) (*source.Release, error) {
	r, _, err := s.client.Releases.GetRelease(s.repo, version, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return s.parseRelease(ctx, r), nil
}

func (s *gitlabSource) parseRelease(ctx context.Context, release *gitlab.Release) *source.Release {
	r := &source.Release{
		Name:        release.Name,
		Description: release.Description,
		Version:     release.TagName,
		Date:        *release.CreatedAt,
		Assets:      make([]*source.Asset, 0, len(release.Assets.Links)),
	}

	for _, l := range release.Assets.Links {
		size, err := s.getSize(ctx, l.URL)
		if err != nil {
			log.Printf("failed to get size for %s: %s\n", l.Name, err)
		}

		r.Assets = append(r.Assets, &source.Asset{
			Name: l.Name,
			URL:  l.URL,
			Size: size,
		})
	}

	return r
}

func (s *gitlabSource) UploadAsset(ctx context.Context, version, name string, data []byte) error {
	file, _, err := s.client.Projects.UploadFile(s.repo, bytes.NewReader(data), name, gitlab.WithContext(ctx))
	if err != nil {
		return err
	}

	u := s.client.BaseURL()
	u.Path = path.Join(s.repo, file.URL)
	url := u.String()

	opt := &gitlab.CreateReleaseLinkOptions{Name: &name, URL: &url}
	_, _, err = s.client.ReleaseLinks.CreateReleaseLink(s.repo, version, opt, gitlab.WithContext(ctx))

	return err
}

func (s *gitlabSource) DownloadAsset(ctx context.Context, version, name string) ([]byte, error) {
	links, _, err := s.client.ReleaseLinks.ListReleaseLinks(s.repo, version, nil)
	if err != nil {
		return nil, err
	}

	for _, link := range links {
		if link.Name == name {
			req, err := retryablehttp.NewRequest(http.MethodGet, link.URL, nil)
			if err != nil {
				return nil, err
			}

			var buf bytes.Buffer
			_, err = s.client.Do(req.WithContext(ctx), &buf)
			if err != nil {
				return nil, err
			}

			return buf.Bytes(), nil
		}
	}

	return nil, source.ErrAssetNotFound
}

func (s *gitlabSource) getSize(ctx context.Context, url string) (int, error) {
	req, err := retryablehttp.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return 0, err
	}

	r, err := s.client.Do(req.WithContext(ctx), nil)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(r.Header.Get("Content-Length"))
}

//nolint:gochecknoinits
func init() { source.Register("gitlab", New) }
