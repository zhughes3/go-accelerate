package slog

type fieldLabels struct {
	duration string
	error    string
	function string
	source   string
}

var (
	defaultLabels = fieldLabels{
		duration: "Duration",
		error:    "Error",
		function: "Function",
		source:   "Source",
	}

	lowercaseLabels = fieldLabels{
		duration: "duration",
		error:    "error",
		function: "function",
		source:   "source",
	}
)
