package version

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func Set(v, c, d string) {
	Version = v
	Commit = c
	Date = d
}
