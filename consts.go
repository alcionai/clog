package clog

const clogFileEnv = "CLOG_FILE"

type logLevel string

const (
	LLDebug    logLevel = "debug"
	LLInfo     logLevel = "info"
	LLError    logLevel = "error"
	LLDisabled logLevel = "disabled"
)

type logFormat string

const (
	// use for cli/terminal
	LFHuman logFormat = "human"
	// use for cloud logging
	LFJSON logFormat = "json"
)

type piiAlg string

const (
	PIIHash      piiAlg = "hash"
	PIIMask      piiAlg = "mask"
	PIIPlainText piiAlg = "plaintext"
)

// FIXME: use a config interface instead; no flag values here.
//
// flag names
const (
	DebugAPIFN          = "debug-api-calls"
	LogFileFN           = "log-file"
	LogFormatFN         = "log-format"
	LogLevelFN          = "log-level"
	MaskSensitiveDataFN = "mask-sensitive-data"
)

// flag values
var (
	DebugAPIFV          bool
	logFileFV           string
	LogFormatFV         string
	LogLevelFV          string
	MaskSensitiveDataFV bool

	ResolvedLogFile string // logFileFV after processing
	piiHandling     string // piiHandling after MaskSensitiveDataFV processing
)

const (
	Stderr = "stderr"
	Stdout = "stdout"
)
