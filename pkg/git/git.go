package git

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

const (
	ShortHashLen = 7
)

type Repository struct {
	*git.Repository

	tags map[plumbing.Hash]*plumbing.Reference
}

func Open(dir string) (*Repository, error) {
	r, err := git.PlainOpen(dir)
	if err != nil {
		return nil, err
	}
	return &Repository{
		Repository: r,
		tags:       make(map[plumbing.Hash]*plumbing.Reference),
	}, err
}

func (r *Repository) getTags() error {
	tags, err := r.Tags()
	if err != nil {
		return err
	}
	err = tags.ForEach(func(t *plumbing.Reference) error {
		r.tags[t.Hash()] = t
		return nil
	})
	return err
}

func (r *Repository) RemoteURL(name string) (string, error) {
	remote, err := r.Remote(name)
	if err != nil {
		return "", err
	}
	if len(remote.Config().URLs) > 0 {
		return remote.Config().URLs[0], nil
	}
	return "", nil
}

type TreeState string

const (
	GitTreeClean TreeState = "clean"
	GitTreeDirty TreeState = "dirty"
)

func (r *Repository) TreeState() (TreeState, error) {
	wt, err := r.Repository.Worktree()
	if err != nil {
		return "", err
	}

	status, err := wt.Status()
	if err != nil {
		return "", err
	}

	if len(status) > 0 {
		return GitTreeDirty, nil
	}
	return GitTreeClean, nil
}

// Describe the reference as 'git describe --tags' will do
func (r *Repository) Describe(ref *plumbing.Reference) (*DescribeObject, error) {
	if ref == nil {
		var err error
		ref, err = r.Head()
		if err != nil {
			return nil, err
		}
	}

	err := r.getTags()
	if err != nil {
		return nil, err
	}

	if len(r.tags) == 0 {
		return &DescribeObject{
			Tag:   nil,
			Count: 0,
			Hash:  ref.Hash(),
		}, nil
	}

	cIter, err := r.Log(&git.LogOptions{
		From:  ref.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, err
	}

	var tag *plumbing.Reference
	var count int
	err = cIter.ForEach(func(c *object.Commit) error {
		if t, ok := r.tags[c.Hash]; ok {
			tag = t
		}
		if tag != nil {
			return storer.ErrStop
		}
		count++
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &DescribeObject{
		Tag:   tag,
		Count: count,
		Hash:  ref.Hash(),
	}, nil
}

type DescribeObject struct {
	Tag   *plumbing.Reference
	Count int
	Hash  plumbing.Hash
}

// This translates the "git describe" to an actual semver.org
// compatible semantic version that looks something like this:
//
//	v1.1.0-alpha.0.6+84c76d1142ea4d
func (desc *DescribeObject) SemanticVersion() string {
	version := "v0.0.0"
	hash := desc.Hash.String()[0:ShortHashLen]
	if desc.Tag == nil {
		if desc.Count > 0 {
			return fmt.Sprintf("%s-%v+%s", version, desc.Count, hash)
		}
		return fmt.Sprintf("%s-%s", version, hash)
	}

	version = desc.Tag.Name().Short()
	if desc.Count == 0 {
		return version
	}
	if strings.Contains(version, "-") {
		return fmt.Sprintf("%s.%v+%v", version, desc.Count, hash)
	}
	return fmt.Sprintf("%s-%v+%s", version, desc.Count, hash)
}

// docker tag cann't contain '+'
func (desc *DescribeObject) DokcerTag() string {
	version := "v0.0.0"
	hash := desc.Hash.String()[0:ShortHashLen]
	if desc.Tag == nil {
		if desc.Count > 0 {
			return fmt.Sprintf("%s-%v-%s", version, desc.Count, hash)
		}
		return fmt.Sprintf("%s-%s", version, hash)
	}

	version = desc.Tag.Name().Short()
	if desc.Count == 0 {
		return version
	}
	if strings.Contains(version, "-") {
		return fmt.Sprintf("%s.%v-%v", version, desc.Count, hash)
	}
	return fmt.Sprintf("%s-%v-%s", version, desc.Count, hash)
}
