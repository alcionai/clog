package clog

import (
	"os"
	"path/filepath"
	"time"

	"golang.org/x/exp/slices"

	"github.com/alcionai/canario/src/internal/common/str"
	"github.com/alcionai/clues"
)

// Settings records the user's preferred logging settings.
type Settings struct {
	// core settings
	File   string    // what file to log to (alt: stderr, stdout)
	Format logFormat // whether to format as text (console) or json (cloud)
	Level  logLevel  // what level to log at

	// more fiddly bits
	PIIHandling piiAlg // how to obscure pii
}

// EnsureDefaults sets any non-populated settings to their default value.
// exported for testing without circular dependencies.
func (s Settings) EnsureDefaults() Settings {
	set := s

	levels := []logLevel{LLDisabled, LLDebug, LLInfo, LLError}
	if len(set.Level) == 0 || !slices.Contains(levels, set.Level) {
		set.Level = LLInfo
	}

	formats := []logFormat{LFText, LFJSON}
	if len(set.Format) == 0 || !slices.Contains(formats, set.Format) {
		set.Format = LFText
	}

	algs := []piiAlg{PIIPlainText, PIIMask, PIIHash}
	if len(set.PIIHandling) == 0 || !slices.Contains(algs, set.PIIHandling) {
		set.PIIHandling = piiAlg(str.First(piiHandling, string(PIIPlainText)))
	}

	if len(set.File) == 0 {
		set.File = GetLogFile("")
		ResolvedLogFile = set.File
	}

	return set
}

// Returns the default location for log file storage.
func defaultLogLocation() string {
	return filepath.Join(
		userLogsDir,
		"clog",
		time.Now().UTC().Format("2006-01-02T15-04-05Z")+".log")
}

// GetLogFile finds the log file in the users local system.
// Uses the env var declaration, if populated, else defaults to stderr.
func GetLogFile(logFileFlagVal string) string {
	if len(ResolvedLogFile) > 0 {
		return ResolvedLogFile
	}

	r := os.Getenv(clogFileEnv)

	// if no env var is specified, fall back to the default file location.
	if len(r) == 0 {
		r = defaultLogLocation()
	}

	// direct to Stdout if provided '-'.
	if r == "-" {
		r = Stdout
	}

	// if outputting to a file, make sure we can access the file.
	if r != Stdout && r != Stderr {
		logdir := filepath.Dir(r)

		err := os.MkdirAll(logdir, 0o755)
		if err != nil {
			return Stderr
		}
	}

	return r
}

func setCluesSecretsHash(alg piiAlg) {
	switch alg {
	case PIIHash:
		clues.SetHasher(clues.DefaultHash())
	case PIIMask:
		clues.SetHasher(clues.HashCfg{HashAlg: clues.Flatmask})
	case PIIPlainText:
		clues.SetHasher(clues.NoHash())
	}
}
