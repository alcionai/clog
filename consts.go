package clog

// Default / Example labels
const (
	// good for categorizing debugging-level api call info.
	LAPICall = "clabel_api_call"
	// when you want your log to cause a lot of noise.
	LAlarmOnThis = "clabel_alarm_on_this"
	// for info about end-of-run resource cleanup.
	LCleanup = "clabel_cleanup"
	// for showcasing the runtime configuration of your app
	LConfiguration = "clabel_configuration"
	// everything that you want to know about the process
	// at the time of its conclusion.
	LEndOfRun = "clabel_end_of_run"
	// good for marking the the error logs that you need to review
	// when debugging "what exactly failed in this run?"
	LFailureOrigin = "clabel_failure_origin"
	// when you want debug logging to include info about every item
	// that gets handled through the process.
	LIndividualItemDetails = "clabel_individual_item_details"
	// when debugging the progress of a process and you want to
	// include logs that track the completion of long running
	// processes.
	LProgressTicker = "clabel_progress_ticker"
	// everything that you want to know about the state of the
	// application when you kick off a new process.
	LStartOfRun = "clabel_start_of_run"
	// who needs a logging level when you can use a label instead?
	LWarning = "clabel_warning"
)
