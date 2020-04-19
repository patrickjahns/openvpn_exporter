package version

import (
	"runtime"
	"time"
)

var (
	// String gets defined by the build system.
	String = "0.0.0"

	// Revision gets defined by the built system
	Revision = ""

	// Date defines the date this binary was built.
	Date string

	// Go running this binary.
	Go = runtime.Version()

	// Started has the time this was started.
	Started = time.Now()
)
