package git

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
)

func TestDescribeObject_SemanticVersion(t *testing.T) {
	tests := []struct {
		name string
		desc DescribeObject
		want string
	}{
		{
			"no tag",
			DescribeObject{
				Count: 0,
				Hash:  plumbing.NewHash("edfd80ededf0e93ebc64ac9421807da5b3d7facf"),
			},
			"v0.0.0-edfd80e",
		},
		{
			"no tag",
			DescribeObject{
				Count: 1,
				Hash:  plumbing.NewHash("edfd80ededf0e93ebc64ac9421807da5b3d7facf"),
			},
			"v0.0.0-1+edfd80e",
		},
		{
			"",
			DescribeObject{
				Tag:   plumbing.NewHashReference(plumbing.NewTagReferenceName("v0.1.0"), plumbing.NewHash("0000000")),
				Count: 0,
				Hash:  plumbing.NewHash("edfd80ededf0e93ebc64ac9421807da5b3d7facf"),
			},
			"v0.1.0",
		},
		{
			"",
			DescribeObject{
				Tag:   plumbing.NewHashReference(plumbing.NewTagReferenceName("v0.2.0"), plumbing.NewHash("0000000")),
				Count: 1,
				Hash:  plumbing.NewHash("edfd80ededf0e93ebc64ac9421807da5b3d7facf"),
			},
			"v0.2.0-1+edfd80e",
		},
		{
			"",
			DescribeObject{
				Tag:   plumbing.NewHashReference(plumbing.NewTagReferenceName("v1.0.0-alpha"), plumbing.NewHash("0000000")),
				Count: 1,
				Hash:  plumbing.NewHash("edfd80ededf0e93ebc64ac9421807da5b3d7facf"),
			},
			"v1.0.0-alpha.1+edfd80e",
		},
		// TODO: Add test cases.
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.desc.SemanticVersion(); got != tt.want {
				t.Errorf("DescribeObject.SemanticVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDescribeObject_DockerTag(t *testing.T) {
	tests := []struct {
		name string
		desc DescribeObject
		want string
	}{
		{
			"no tag",
			DescribeObject{
				Count: 0,
				Hash:  plumbing.NewHash("edfd80ededf0e93ebc64ac9421807da5b3d7facf"),
			},
			"v0.0.0-edfd80e",
		},
		{
			"no tag",
			DescribeObject{
				Count: 1,
				Hash:  plumbing.NewHash("edfd80ededf0e93ebc64ac9421807da5b3d7facf"),
			},
			"v0.0.0-1-edfd80e",
		},
		{
			"",
			DescribeObject{
				Tag:   plumbing.NewHashReference(plumbing.NewTagReferenceName("v0.1.0"), plumbing.NewHash("0000000")),
				Count: 0,
				Hash:  plumbing.NewHash("edfd80ededf0e93ebc64ac9421807da5b3d7facf"),
			},
			"v0.1.0",
		},
		{
			"",
			DescribeObject{
				Tag:   plumbing.NewHashReference(plumbing.NewTagReferenceName("v0.2.0"), plumbing.NewHash("0000000")),
				Count: 1,
				Hash:  plumbing.NewHash("edfd80ededf0e93ebc64ac9421807da5b3d7facf"),
			},
			"v0.2.0-1-edfd80e",
		},
		{
			"",
			DescribeObject{
				Tag:   plumbing.NewHashReference(plumbing.NewTagReferenceName("v1.0.0-alpha"), plumbing.NewHash("0000000")),
				Count: 1,
				Hash:  plumbing.NewHash("edfd80ededf0e93ebc64ac9421807da5b3d7facf"),
			},
			"v1.0.0-alpha.1-edfd80e",
		},
		// TODO: Add test cases.
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.desc.DokcerTag(); got != tt.want {
				t.Errorf("DescribeObject.DockerTag() = %v, want %v", got, tt.want)
			}
		})
	}
}
