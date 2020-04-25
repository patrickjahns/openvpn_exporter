package version

import (
	"fmt"
	"runtime"
	"time"
)

var (
	// Version gets defined by the build system.
	Version = "0.0.0"

	// Revision gets defined by the built system
	Revision = ""

	// BuildDate defines the date this binary was built.
	BuildDate string

	// GoVersion running this binary.
	GoVersion = runtime.Version()

	// Started has the time this was started.
	Started = time.Now()
)

// Info returns version, revision information.
func Info() string {
	return fmt.Sprintf("%s (%s)", Version, Revision)
}

// BuildContext returns goVersion, buildUser and buildDate information.
func BuildContext() string {
	return fmt.Sprintf("(go=%s, date=%s)", GoVersion, BuildDate)
}
