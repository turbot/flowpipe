package version

// This is the default version that will be used if the linker flags are not supplied during build time
//
// # GitHub action builds are setup with the correct linker flags
//
// Local build (Makefile): use the default value in DOCKERFILE

// TODO KAI unify this with steampipe
var (
	version   = "0.0.1-beta.1"
	buildTime string
	commit    string
	builtBy   string
)

func GetVersion() string {
	return version
}

func GetBuildTime() string {
	return buildTime
}

func GetCommit() string {
	return commit
}

func GetBuiltBy() string {
	return builtBy
}
