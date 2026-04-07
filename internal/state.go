package internal

var (
	// NOTE(d1): global
	Debug   bool
	NoInput bool
	Offline bool
	Help    bool
	Version bool

	// NOTE(d1): sub-command specific
	Chaos            bool
	DeployLatest     bool
	DontWaitConverge bool
	Dry              bool
	Force            bool
	Secrets  		 bool
	Major            bool
	Minor            bool
	NoDomainChecks   bool
	Patch            bool
	ShowUnchanged    bool
)