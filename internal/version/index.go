package version

// This is the default version that will be used if the linker flags are not supplied during build time
//
// # GitHub action builds are setup with the correct linker flags
//
// Local build (Makefile): use the default value in DOCKERFILE
var Version = "0.0.1-beta.1"

var BuildTime string

var Commit string

var BuiltBy string

func GetVersion() string {
	return Version
}

func GetBuildTime() string {
	return BuildTime
}

func GetCommit() string {
	return Commit
}

func GetBuiltBy() string {
	return BuiltBy
}
