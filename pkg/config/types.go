package config

// Config is the unmarshalled representation of the configuration file
type Config struct {
	// Version is the project version, defaults to "1" (backwards compatibility)
	Version string `json:"version,omitempty"`

	// Go config
	Go Go `json:"go,omitempty"`

	// container config
	Container Container `json:"container,omitempty"`
}

type Go struct {
	Build  GoBuild  `json:"build,omitempty"`
	Mod    GoMod    `json:"mod,omitempty"`
	Format GoFormat `json:"format,omitempty"`
}

type GoBuild struct {
	Platforms      []string `json:"platforms,omitempty"`
	OnBuildImage   string   `json:"onBuildImage,omitempty"`
	GlobalHooksDir string   `json:"globalHooksDir,omitempty"`
}

type GoMod struct {
	Require []GoModRequire `json:"require,omitempty"`
	Replace []GoModReplace `json:"replace,omitempty"`
}

type GoModRequire struct {
	Path     string `json:"path,omitempty"`
	Version  string `json:"version,omitempty"`
	SkipDeps bool   `json:"skipDeps,omitempty"`
}

type GoModReplace struct {
	OldPath string `json:"path,omitempty"`
	NewPath string `json:"newPath,omitempty"`
	Version string `json:"version,omitempty"`
}

type GoFormat struct {
	Local   string          `json:"local,omitempty"`
	Exclude GoFormatExclude `json:"exclude,omitempty"`
}

type GoFormatExclude struct {
	Dirs  []string `json:"dirs,omitempty"`
	Files []string `json:"files,omitempty"`
}

type Container struct {
	Registries  []string `json:"registries,omitempty"`
	ImagePrefix string   `json:"imagePrefix,omitempty"`
	ImageSuffix string   `json:"imageSuffix,omitempty"`
}
