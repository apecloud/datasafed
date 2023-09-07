package version

// Version is the string that contains version
var Version = "edge"

// BuildDate is the string of binary build date
var BuildDate string

// GitCommit is the string of git commit ID
var GitCommit string

// GitVersion is the string of git version tag
var GitVersion string

// GetVersion returns the version for cli, either got from "git describe --tags" or "dev" when doing simple go build
func GetVersion() string {
	if len(Version) == 0 {
		return "v1-dev"
	}
	return Version
}
