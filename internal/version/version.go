package version

// Set at build time via ldflags:
// -X github.com/smsufyian/sprig/internal/version.Version=1.0.0
// -X github.com/smsufyian/sprig/internal/version.Commit=abc1234
// -X github.com/smsufyian/sprig/internal/version.BuildDate=2026-05-10T09:00:00Z
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)
