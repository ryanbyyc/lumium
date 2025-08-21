package svckit

// Kit holds shared deps for a service: DB, Repo, and Cfg
// Embed *Kit[...] in concrete service type
type Kit[DB any, Repo any, Cfg any] struct {
	DB   DB
	Repo Repo
	Cfg  Cfg
}

// Opt is an option that mutates the kit at construction time
type Opt[DB any, Repo any, Cfg any] func(*Kit[DB, Repo, Cfg])

// WithRepo lets you override the default repo (e.g., a mock in tests)
func WithRepo[DB any, Repo any, Cfg any](r Repo) Opt[DB, Repo, Cfg] {
	return func(k *Kit[DB, Repo, Cfg]) { k.Repo = r }
}

// New constructs a Kit with a default repo via repoCtor, then applies opts
func New[DB any, Repo any, Cfg any](
	db DB,
	repoCtor func() Repo,
	cfg Cfg,
	opts ...Opt[DB, Repo, Cfg],
) *Kit[DB, Repo, Cfg] {
	k := &Kit[DB, Repo, Cfg]{DB: db, Repo: repoCtor(), Cfg: cfg}
	for _, o := range opts {
		o(k)
	}
	return k
}