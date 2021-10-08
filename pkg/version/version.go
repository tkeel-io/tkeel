package version

// Values for these are injected by the build.
var (
	version = "edge"

	gitcommit, gitversion string
)

// Version returns the Keel version. This is either a semantic version
// number or else, in the case of unreleased code, the string "edge".
func Version() string {
	return version
}

// Commit returns the git commit SHA for the code that Keel was built from.
func Commit() string {
	return gitcommit
}

// GitVersion returns the git version for the code that Keel was built from.
func GitVersion() string {
	return gitversion
}
