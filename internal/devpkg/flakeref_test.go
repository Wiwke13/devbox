package devpkg

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseFlakeRef(t *testing.T) {
	cases := map[string]FlakeRef{
		// Path-like references start with a '.' or '/'.
		// This distinguishes them from indirect references
		// (./nixpkgs is a directory; nixpkgs is an indirect).
		".":                {Type: FlakeTypePath, Path: "."},
		"./":               {Type: FlakeTypePath, Path: "./"},
		"./flake":          {Type: FlakeTypePath, Path: "./flake"},
		"./relative/flake": {Type: FlakeTypePath, Path: "./relative/flake"},
		"/":                {Type: FlakeTypePath, Path: "/"},
		"/flake":           {Type: FlakeTypePath, Path: "/flake"},
		"/absolute/flake":  {Type: FlakeTypePath, Path: "/absolute/flake"},

		// Path-like references can have raw unicode characters unlike
		// path: URL references.
		"./Ûñî©ôδ€/flake\n": {Type: FlakeTypePath, Path: "./Ûñî©ôδ€/flake\n"},
		"/Ûñî©ôδ€/flake\n":  {Type: FlakeTypePath, Path: "/Ûñî©ôδ€/flake\n"},

		// URL-like path references.
		"path:":                      {Type: FlakeTypePath, Path: ""},
		"path:.":                     {Type: FlakeTypePath, Path: "."},
		"path:./":                    {Type: FlakeTypePath, Path: "./"},
		"path:./flake":               {Type: FlakeTypePath, Path: "./flake"},
		"path:./relative/flake":      {Type: FlakeTypePath, Path: "./relative/flake"},
		"path:./relative/my%20flake": {Type: FlakeTypePath, Path: "./relative/my flake"},
		"path:/":                     {Type: FlakeTypePath, Path: "/"},
		"path:/flake":                {Type: FlakeTypePath, Path: "/flake"},
		"path:/absolute/flake":       {Type: FlakeTypePath, Path: "/absolute/flake"},

		// URL-like paths can omit the "./" prefix for relative
		// directories.
		"path:flake":          {Type: FlakeTypePath, Path: "flake"},
		"path:relative/flake": {Type: FlakeTypePath, Path: "relative/flake"},

		// Indirect references.
		"flake:indirect":          {Type: FlakeTypeIndirect, ID: "indirect"},
		"flake:indirect/ref":      {Type: FlakeTypeIndirect, ID: "indirect", Ref: "ref"},
		"flake:indirect/my%2Fref": {Type: FlakeTypeIndirect, ID: "indirect", Ref: "my/ref"},
		"flake:indirect/5233fd2ba76a3accb5aaa999c00509a11fd0793c":     {Type: FlakeTypeIndirect, ID: "indirect", Rev: "5233fd2ba76a3accb5aaa999c00509a11fd0793c"},
		"flake:indirect/ref/5233fd2ba76a3accb5aaa999c00509a11fd0793c": {Type: FlakeTypeIndirect, ID: "indirect", Ref: "ref", Rev: "5233fd2ba76a3accb5aaa999c00509a11fd0793c"},

		// Indirect references can omit their "indirect:" type prefix.
		"indirect":     {Type: FlakeTypeIndirect, ID: "indirect"},
		"indirect/ref": {Type: FlakeTypeIndirect, ID: "indirect", Ref: "ref"},
		"indirect/5233fd2ba76a3accb5aaa999c00509a11fd0793c":     {Type: FlakeTypeIndirect, ID: "indirect", Rev: "5233fd2ba76a3accb5aaa999c00509a11fd0793c"},
		"indirect/ref/5233fd2ba76a3accb5aaa999c00509a11fd0793c": {Type: FlakeTypeIndirect, ID: "indirect", Ref: "ref", Rev: "5233fd2ba76a3accb5aaa999c00509a11fd0793c"},

		// GitHub references.
		"github:NixOS/nix":            {Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix"},
		"github:NixOS/nix/v1.2.3":     {Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Ref: "v1.2.3"},
		"github:NixOS/nix?ref=v1.2.3": {Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Ref: "v1.2.3"},
		"github:NixOS/nix?ref=5233fd2ba76a3accb5aaa999c00509a11fd0793c": {Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Ref: "5233fd2ba76a3accb5aaa999c00509a11fd0793c"},
		"github:NixOS/nix/5233fd2ba76a3accb5aaa999c00509a11fd0793c":     {Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Rev: "5233fd2ba76a3accb5aaa999c00509a11fd0793c"},
		"github:NixOS/nix/5233fd2bb76a3accb5aaa999c00509a11fd0793z":     {Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Ref: "5233fd2bb76a3accb5aaa999c00509a11fd0793z"},
		"github:NixOS/nix?rev=5233fd2ba76a3accb5aaa999c00509a11fd0793c": {Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Rev: "5233fd2ba76a3accb5aaa999c00509a11fd0793c"},
		"github:NixOS/nix?host=example.com":                             {Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Host: "example.com"},

		// The github type allows clone-style URLs. The username and
		// host are ignored.
		"github://git@github.com/NixOS/nix":                                              {Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix"},
		"github://git@github.com/NixOS/nix/v1.2.3":                                       {Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Ref: "v1.2.3"},
		"github://git@github.com/NixOS/nix?ref=v1.2.3":                                   {Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Ref: "v1.2.3"},
		"github://git@github.com/NixOS/nix?ref=5233fd2ba76a3accb5aaa999c00509a11fd0793c": {Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Ref: "5233fd2ba76a3accb5aaa999c00509a11fd0793c"},
		"github://git@github.com/NixOS/nix?rev=5233fd2ba76a3accb5aaa999c00509a11fd0793c": {Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Rev: "5233fd2ba76a3accb5aaa999c00509a11fd0793c"},
		"github://git@github.com/NixOS/nix?host=example.com":                             {Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Host: "example.com"},

		// Git references.
		"git://example.com/repo/flake":         {Type: FlakeTypeGit, URL: "git://example.com/repo/flake"},
		"git+https://example.com/repo/flake":   {Type: FlakeTypeGit, URL: "https://example.com/repo/flake"},
		"git+ssh://git@example.com/repo/flake": {Type: FlakeTypeGit, URL: "ssh://git@example.com/repo/flake"},
		"git:/repo/flake":                      {Type: FlakeTypeGit, URL: "git:/repo/flake"},
		"git+file:///repo/flake":               {Type: FlakeTypeGit, URL: "file:///repo/flake"},
		"git://example.com/repo/flake?ref=unstable&rev=e486d8d40e626a20e06d792db8cc5ac5aba9a5b4&dir=subdir": {Type: FlakeTypeGit, URL: "git://example.com/repo/flake?dir=subdir", Ref: "unstable", Rev: "e486d8d40e626a20e06d792db8cc5ac5aba9a5b4", Dir: "subdir"},

		// Tarball references.
		"tarball+http://example.com/flake":  {Type: FlakeTypeTarball, URL: "http://example.com/flake"},
		"tarball+https://example.com/flake": {Type: FlakeTypeTarball, URL: "https://example.com/flake"},
		"tarball+file:///home/flake":        {Type: FlakeTypeTarball, URL: "file:///home/flake"},

		// Regular URLs have the tarball type if they have a known
		// archive extension:
		// .zip, .tar, .tgz, .tar.gz, .tar.xz, .tar.bz2 or .tar.zst
		"http://example.com/flake.zip":            {Type: FlakeTypeTarball, URL: "http://example.com/flake.zip"},
		"http://example.com/flake.tar":            {Type: FlakeTypeTarball, URL: "http://example.com/flake.tar"},
		"http://example.com/flake.tgz":            {Type: FlakeTypeTarball, URL: "http://example.com/flake.tgz"},
		"http://example.com/flake.tar.gz":         {Type: FlakeTypeTarball, URL: "http://example.com/flake.tar.gz"},
		"http://example.com/flake.tar.xz":         {Type: FlakeTypeTarball, URL: "http://example.com/flake.tar.xz"},
		"http://example.com/flake.tar.bz2":        {Type: FlakeTypeTarball, URL: "http://example.com/flake.tar.bz2"},
		"http://example.com/flake.tar.zst":        {Type: FlakeTypeTarball, URL: "http://example.com/flake.tar.zst"},
		"http://example.com/flake.tar?dir=subdir": {Type: FlakeTypeTarball, URL: "http://example.com/flake.tar?dir=subdir", Dir: "subdir"},
		"file:///flake.zip":                       {Type: FlakeTypeTarball, URL: "file:///flake.zip"},
		"file:///flake.tar":                       {Type: FlakeTypeTarball, URL: "file:///flake.tar"},
		"file:///flake.tgz":                       {Type: FlakeTypeTarball, URL: "file:///flake.tgz"},
		"file:///flake.tar.gz":                    {Type: FlakeTypeTarball, URL: "file:///flake.tar.gz"},
		"file:///flake.tar.xz":                    {Type: FlakeTypeTarball, URL: "file:///flake.tar.xz"},
		"file:///flake.tar.bz2":                   {Type: FlakeTypeTarball, URL: "file:///flake.tar.bz2"},
		"file:///flake.tar.zst":                   {Type: FlakeTypeTarball, URL: "file:///flake.tar.zst"},
		"file:///flake.tar?dir=subdir":            {Type: FlakeTypeTarball, URL: "file:///flake.tar?dir=subdir", Dir: "subdir"},

		// File URL references.
		"file+file:///flake":                           {Type: FlakeTypeFile, URL: "file:///flake"},
		"file+http://example.com/flake":                {Type: FlakeTypeFile, URL: "http://example.com/flake"},
		"file+http://example.com/flake.git":            {Type: FlakeTypeFile, URL: "http://example.com/flake.git"},
		"file+http://example.com/flake.tar?dir=subdir": {Type: FlakeTypeFile, URL: "http://example.com/flake.tar?dir=subdir", Dir: "subdir"},

		// Regular URLs have the file type if they don't have a known
		// archive extension.
		"http://example.com/flake":            {Type: FlakeTypeFile, URL: "http://example.com/flake"},
		"http://example.com/flake.git":        {Type: FlakeTypeFile, URL: "http://example.com/flake.git"},
		"http://example.com/flake?dir=subdir": {Type: FlakeTypeFile, URL: "http://example.com/flake?dir=subdir", Dir: "subdir"},
	}
	for ref, want := range cases {
		t.Run(ref, func(t *testing.T) {
			got, err := ParseFlakeRef(ref)
			if diff := cmp.Diff(want, got); diff != "" {
				if err != nil {
					t.Errorf("got error: %s", err)
				}
				t.Errorf("wrong flakeref (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseFlakeRefError(t *testing.T) {
	t.Run("EmptyString", func(t *testing.T) {
		ref := ""
		_, err := ParseFlakeRef(ref)
		if err == nil {
			t.Error("got nil error for bad flakeref:", ref)
		}
	})
	t.Run("InvalidURL", func(t *testing.T) {
		ref := "://bad/url"
		_, err := ParseFlakeRef(ref)
		if err == nil {
			t.Error("got nil error for bad flakeref:", ref)
		}
	})
	t.Run("InvalidURLEscape", func(t *testing.T) {
		ref := "path:./relative/my%flake"
		_, err := ParseFlakeRef(ref)
		if err == nil {
			t.Error("got nil error for bad flakeref:", ref)
		}
	})
	t.Run("UnsupportedURLScheme", func(t *testing.T) {
		ref := "runx:mvdan/gofumpt@latest"
		_, err := ParseFlakeRef(ref)
		if err == nil {
			t.Error("got nil error for bad flakeref:", ref)
		}
	})
	t.Run("PathLikeWith?#", func(t *testing.T) {
		in := []string{
			"./invalid#path",
			"./invalid?path",
			"/invalid#path",
			"/invalid?path",
			"/#",
			"/?",
		}
		for _, ref := range in {
			_, err := ParseFlakeRef(ref)
			if err == nil {
				t.Error("got nil error for bad flakeref:", ref)
			}
		}
	})
	t.Run("GitHubInvalidRefRevCombo", func(t *testing.T) {
		in := []string{
			"github:NixOS/nix?ref=v1.2.3&rev=5233fd2ba76a3accb5aaa999c00509a11fd0793c",
			"github:NixOS/nix/v1.2.3?ref=v4.5.6",
			"github:NixOS/nix/5233fd2ba76a3accb5aaa999c00509a11fd0793c?rev=e486d8d40e626a20e06d792db8cc5ac5aba9a5b4",
			"github:NixOS/nix/5233fd2ba76a3accb5aaa999c00509a11fd0793c?ref=v1.2.3",
		}
		for _, ref := range in {
			_, err := ParseFlakeRef(ref)
			if err == nil {
				t.Error("got nil error for bad flakeref:", ref)
			}
		}
	})
}

func TestFlakeRefString(t *testing.T) {
	cases := map[FlakeRef]string{
		{}: "",

		// Path references.
		{Type: FlakeTypePath, Path: "."}:                "path:.",
		{Type: FlakeTypePath, Path: "./"}:               "path:.",
		{Type: FlakeTypePath, Path: "./flake"}:          "path:flake",
		{Type: FlakeTypePath, Path: "./relative/flake"}: "path:relative/flake",
		{Type: FlakeTypePath, Path: "/"}:                "path:/",
		{Type: FlakeTypePath, Path: "/flake"}:           "path:/flake",
		{Type: FlakeTypePath, Path: "/absolute/flake"}:  "path:/absolute/flake",

		// Path references with escapes.
		{Type: FlakeTypePath, Path: "%"}:                 "path:%25",
		{Type: FlakeTypePath, Path: "/%2F"}:              "path:/%252F",
		{Type: FlakeTypePath, Path: "./Ûñî©ôδ€/flake\n"}: "path:%C3%9B%C3%B1%C3%AE%C2%A9%C3%B4%CE%B4%E2%82%AC/flake%0A",
		{Type: FlakeTypePath, Path: "/Ûñî©ôδ€/flake\n"}:  "path:/%C3%9B%C3%B1%C3%AE%C2%A9%C3%B4%CE%B4%E2%82%AC/flake%0A",

		// Indirect references.
		{Type: FlakeTypeIndirect, ID: "indirect"}:                                                              "flake:indirect",
		{Type: FlakeTypeIndirect, ID: "indirect", Dir: "sub/dir"}:                                              "flake:indirect?dir=sub%2Fdir",
		{Type: FlakeTypeIndirect, ID: "indirect", Ref: "ref"}:                                                  "flake:indirect/ref",
		{Type: FlakeTypeIndirect, ID: "indirect", Ref: "my/ref"}:                                               "flake:indirect/my%2Fref",
		{Type: FlakeTypeIndirect, ID: "indirect", Rev: "5233fd2ba76a3accb5aaa999c00509a11fd0793c"}:             "flake:indirect/5233fd2ba76a3accb5aaa999c00509a11fd0793c",
		{Type: FlakeTypeIndirect, ID: "indirect", Ref: "ref", Rev: "5233fd2ba76a3accb5aaa999c00509a11fd0793c"}: "flake:indirect/ref/5233fd2ba76a3accb5aaa999c00509a11fd0793c",

		// GitHub references.
		{Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix"}:                                                  "github:NixOS/nix",
		{Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Ref: "v1.2.3"}:                                   "github:NixOS/nix/v1.2.3",
		{Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Ref: "my/ref"}:                                   "github:NixOS/nix/my%2Fref",
		{Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Ref: "5233fd2ba76a3accb5aaa999c00509a11fd0793c"}: "github:NixOS/nix/5233fd2ba76a3accb5aaa999c00509a11fd0793c",
		{Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Ref: "5233fd2bb76a3accb5aaa999c00509a11fd0793z"}: "github:NixOS/nix/5233fd2bb76a3accb5aaa999c00509a11fd0793z",
		{Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Dir: "sub/dir"}:                                  "github:NixOS/nix?dir=sub%2Fdir",
		{Type: FlakeTypeGitHub, Owner: "NixOS", Repo: "nix", Dir: "sub/dir", Host: "example.com"}:             "github:NixOS/nix?dir=sub%2Fdir&host=example.com",

		// Git references.
		{Type: FlakeTypeGit, URL: "git://example.com/repo/flake"}:                                                                     "git://example.com/repo/flake",
		{Type: FlakeTypeGit, URL: "https://example.com/repo/flake"}:                                                                   "git+https://example.com/repo/flake",
		{Type: FlakeTypeGit, URL: "ssh://git@example.com/repo/flake"}:                                                                 "git+ssh://git@example.com/repo/flake",
		{Type: FlakeTypeGit, URL: "git:/repo/flake"}:                                                                                  "git:/repo/flake",
		{Type: FlakeTypeGit, URL: "file:///repo/flake"}:                                                                               "git+file:///repo/flake",
		{Type: FlakeTypeGit, URL: "ssh://git@example.com/repo/flake", Ref: "my/ref", Rev: "e486d8d40e626a20e06d792db8cc5ac5aba9a5b4"}: "git+ssh://git@example.com/repo/flake?ref=my%2Fref&rev=e486d8d40e626a20e06d792db8cc5ac5aba9a5b4",
		{Type: FlakeTypeGit, URL: "ssh://git@example.com/repo/flake?dir=sub%2Fdir", Ref: "my/ref", Rev: "e486d8d40e626a20e06d792db8cc5ac5aba9a5b4", Dir: "sub/dir"}: "git+ssh://git@example.com/repo/flake?dir=sub%2Fdir&ref=my%2Fref&rev=e486d8d40e626a20e06d792db8cc5ac5aba9a5b4",
		{Type: FlakeTypeGit, URL: "git:repo/flake?dir=sub%2Fdir", Ref: "my/ref", Rev: "e486d8d40e626a20e06d792db8cc5ac5aba9a5b4", Dir: "sub/dir"}:                   "git:repo/flake?dir=sub%2Fdir&ref=my%2Fref&rev=e486d8d40e626a20e06d792db8cc5ac5aba9a5b4",

		// Tarball references.
		{Type: FlakeTypeTarball, URL: "http://example.com/flake"}:                  "tarball+http://example.com/flake",
		{Type: FlakeTypeTarball, URL: "https://example.com/flake"}:                 "tarball+https://example.com/flake",
		{Type: FlakeTypeTarball, URL: "https://example.com/flake", Dir: "sub/dir"}: "tarball+https://example.com/flake?dir=sub%2Fdir",
		{Type: FlakeTypeTarball, URL: "file:///home/flake"}:                        "tarball+file:///home/flake",

		// File URL references.
		{Type: FlakeTypeFile, URL: "file:///flake"}:                                              "file+file:///flake",
		{Type: FlakeTypeFile, URL: "http://example.com/flake"}:                                   "file+http://example.com/flake",
		{Type: FlakeTypeFile, URL: "http://example.com/flake.git"}:                               "file+http://example.com/flake.git",
		{Type: FlakeTypeFile, URL: "http://example.com/flake.tar?dir=sub%2Fdir", Dir: "sub/dir"}: "file+http://example.com/flake.tar?dir=sub%2Fdir",
	}

	for ref, want := range cases {
		t.Run(want, func(t *testing.T) {
			t.Logf("input = %#v", ref)
			got := ref.String()
			if got != want {
				t.Errorf("got %#q, want %#q", got, want)
			}
		})
	}
}

func TestParseFlakeInstallable(t *testing.T) {
	cases := map[string]FlakeInstallable{
		// Empty string is not a valid installable.
		"": {},

		// Not a path and not a valid URL.
		"://bad/url": {},

		".":             {Ref: FlakeRef{Type: FlakeTypePath, Path: "."}},
		".#app":         {AttrPath: "app", Ref: FlakeRef{Type: FlakeTypePath, Path: "."}},
		".#app^out":     {AttrPath: "app", Outputs: "out", Ref: FlakeRef{Type: FlakeTypePath, Path: "."}},
		".#app^out,lib": {AttrPath: "app", Outputs: "lib,out", Ref: FlakeRef{Type: FlakeTypePath, Path: "."}},
		".#app^*":       {AttrPath: "app", Outputs: "*", Ref: FlakeRef{Type: FlakeTypePath, Path: "."}},
		".^*":           {Outputs: "*", Ref: FlakeRef{Type: FlakeTypePath, Path: "."}},

		"./flake":             {Ref: FlakeRef{Type: FlakeTypePath, Path: "./flake"}},
		"./flake#app":         {AttrPath: "app", Ref: FlakeRef{Type: FlakeTypePath, Path: "./flake"}},
		"./flake#app^out":     {AttrPath: "app", Outputs: "out", Ref: FlakeRef{Type: FlakeTypePath, Path: "./flake"}},
		"./flake#app^out,lib": {AttrPath: "app", Outputs: "lib,out", Ref: FlakeRef{Type: FlakeTypePath, Path: "./flake"}},
		"./flake^out":         {Outputs: "out", Ref: FlakeRef{Type: FlakeTypePath, Path: "./flake"}},

		"indirect":            {Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "indirect"}},
		"nixpkgs#app":         {AttrPath: "app", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "nixpkgs"}},
		"nixpkgs#app^out":     {AttrPath: "app", Outputs: "out", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "nixpkgs"}},
		"nixpkgs#app^out,lib": {AttrPath: "app", Outputs: "lib,out", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "nixpkgs"}},
		"nixpkgs^out":         {Outputs: "out", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "nixpkgs"}},

		"%23#app":                        {AttrPath: "app", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "#"}},
		"./%23#app":                      {AttrPath: "app", Ref: FlakeRef{Type: FlakeTypePath, Path: "./#"}},
		"/%23#app":                       {AttrPath: "app", Ref: FlakeRef{Type: FlakeTypePath, Path: "/#"}},
		"path:/%23#app":                  {AttrPath: "app", Ref: FlakeRef{Type: FlakeTypePath, Path: "/#"}},
		"http://example.com/%23.tar#app": {AttrPath: "app", Ref: FlakeRef{Type: FlakeTypeTarball, URL: "http://example.com/%23.tar#app"}},
	}

	for installable, want := range cases {
		t.Run(installable, func(t *testing.T) {
			got, err := ParseFlakeInstallable(installable)
			if diff := cmp.Diff(want, got); diff != "" {
				if err != nil {
					t.Errorf("got error: %s", err)
				}
				t.Errorf("wrong installable (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFlakeInstallableString(t *testing.T) {
	cases := map[FlakeInstallable]string{
		{}: "",

		// No attribute or outputs.
		{Ref: FlakeRef{Type: FlakeTypePath, Path: "."}}:          "path:.",
		{Ref: FlakeRef{Type: FlakeTypePath, Path: "./flake"}}:    "path:flake",
		{Ref: FlakeRef{Type: FlakeTypePath, Path: "/flake"}}:     "path:/flake",
		{Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "indirect"}}: "flake:indirect",

		// Attribute without outputs.
		{AttrPath: "app", Ref: FlakeRef{Type: FlakeTypePath, Path: "."}}:            "path:.#app",
		{AttrPath: "my#app", Ref: FlakeRef{Type: FlakeTypePath, Path: "."}}:         "path:.#my%23app",
		{AttrPath: "app", Ref: FlakeRef{Type: FlakeTypePath, Path: "./flake"}}:      "path:flake#app",
		{AttrPath: "my#app", Ref: FlakeRef{Type: FlakeTypePath, Path: "./flake"}}:   "path:flake#my%23app",
		{AttrPath: "app", Ref: FlakeRef{Type: FlakeTypePath, Path: "/flake"}}:       "path:/flake#app",
		{AttrPath: "my#app", Ref: FlakeRef{Type: FlakeTypePath, Path: "/flake"}}:    "path:/flake#my%23app",
		{AttrPath: "app", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "nixpkgs"}}:    "flake:nixpkgs#app",
		{AttrPath: "my#app", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "nixpkgs"}}: "flake:nixpkgs#my%23app",

		// Attribute with single output.
		{AttrPath: "app", Outputs: "out", Ref: FlakeRef{Type: FlakeTypePath, Path: "."}}:         "path:.#app^out",
		{AttrPath: "app", Outputs: "out", Ref: FlakeRef{Type: FlakeTypePath, Path: "./flake"}}:   "path:flake#app^out",
		{AttrPath: "app", Outputs: "out", Ref: FlakeRef{Type: FlakeTypePath, Path: "/flake"}}:    "path:/flake#app^out",
		{AttrPath: "app", Outputs: "out", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "nixpkgs"}}: "flake:nixpkgs#app^out",

		// Attribute with multiple outputs.
		{AttrPath: "app", Outputs: "out,lib", Ref: FlakeRef{Type: FlakeTypePath, Path: "."}}:         "path:.#app^lib,out",
		{AttrPath: "app", Outputs: "out,lib", Ref: FlakeRef{Type: FlakeTypePath, Path: "./flake"}}:   "path:flake#app^lib,out",
		{AttrPath: "app", Outputs: "out,lib", Ref: FlakeRef{Type: FlakeTypePath, Path: "/flake"}}:    "path:/flake#app^lib,out",
		{AttrPath: "app", Outputs: "out,lib", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "nixpkgs"}}: "flake:nixpkgs#app^lib,out",

		// Outputs are cleaned and sorted.
		{AttrPath: "app", Outputs: "out,lib", Ref: FlakeRef{Type: FlakeTypePath, Path: "."}}:       "path:.#app^lib,out",
		{AttrPath: "app", Outputs: "lib,out", Ref: FlakeRef{Type: FlakeTypePath, Path: "./flake"}}: "path:flake#app^lib,out",
		{AttrPath: "app", Outputs: "out,,", Ref: FlakeRef{Type: FlakeTypePath, Path: "/flake"}}:    "path:/flake#app^out",
		{AttrPath: "app", Outputs: ",lib,out", Ref: FlakeRef{Type: FlakeTypePath, Path: "/flake"}}: "path:/flake#app^lib,out",
		{AttrPath: "app", Outputs: ",", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "nixpkgs"}}:     "flake:nixpkgs#app",

		// Wildcard replaces other outputs.
		{AttrPath: "app", Outputs: "*", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "nixpkgs"}}:     "flake:nixpkgs#app^*",
		{AttrPath: "app", Outputs: "out,*", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "nixpkgs"}}: "flake:nixpkgs#app^*",
		{AttrPath: "app", Outputs: ",*", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "nixpkgs"}}:    "flake:nixpkgs#app^*",

		// Outputs are not percent-encoded.
		{AttrPath: "app", Outputs: "%", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "nixpkgs"}}:   "flake:nixpkgs#app^%",
		{AttrPath: "app", Outputs: "/", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "nixpkgs"}}:   "flake:nixpkgs#app^/",
		{AttrPath: "app", Outputs: "%2F", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: "nixpkgs"}}: "flake:nixpkgs#app^%2F",

		// Missing or invalid fields.
		{AttrPath: "app", Ref: FlakeRef{Type: FlakeTypeFile, URL: ""}}:     "",
		{AttrPath: "app", Ref: FlakeRef{Type: FlakeTypeGit, URL: ""}}:      "",
		{AttrPath: "app", Ref: FlakeRef{Type: FlakeTypeGitHub, Owner: ""}}: "",
		{AttrPath: "app", Ref: FlakeRef{Type: FlakeTypeIndirect, ID: ""}}:  "",
		{AttrPath: "app", Ref: FlakeRef{Type: FlakeTypePath, Path: ""}}:    "",
		{AttrPath: "app", Ref: FlakeRef{Type: FlakeTypeTarball, URL: ""}}:  "",
	}

	for installable, want := range cases {
		t.Run(want, func(t *testing.T) {
			t.Logf("input = %#v", installable)
			got := installable.String()
			if got != want {
				t.Errorf("got %#q, want %#q", got, want)
			}
		})
	}
}

func TestBuildQueryString(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("wanted panic for odd-number of key-value parameters")
		}
	}()

	// staticcheck impressively catches buildQueryString calls that have an
	// odd number of parameters. Build the slice in a convoluted way to
	// throw it off and suppress the warning (gopls doesn't have nolint
	// directives).
	var elems []string
	elems = append(elems, "1")
	buildQueryString(elems...)
}
