// Package errors provides exit code constants following sysexits.h conventions.
package errors

// Exit codes following sysexits.h conventions.
// These are standard exit codes used by Unix/Linux systems.
const (
	// ExitOK indicates successful completion.
	ExitOK = 0

	// ExitUsage indicates command line usage error (incorrect arguments).
	ExitUsage = 64

	// ExitDataErr indicates data format error (input data was incorrect).
	ExitDataErr = 65

	// ExitNoInput indicates input file not found or unreadable.
	ExitNoInput = 66

	// ExitNoUser indicates user not found (e.g., authentication failed).
	ExitNoUser = 67

	// ExitNoHost indicates host not found (network error).
	ExitNoHost = 68

	// ExitUnavailable indicates service unavailable (temporary failure).
	ExitUnavailable = 69

	// ExitSoftware indicates internal software error (bug in program).
	ExitSoftware = 70

	// ExitOSErr indicates operating system error (e.g., can't fork).
	ExitOSErr = 71

	// ExitOSFile indicates critical OS file missing (e.g., can't create temp file).
	ExitOSFile = 72

	// ExitCantCreat indicates can't create output file.
	ExitCantCreat = 73

	// ExitIOErr indicates input/output error.
	ExitIOErr = 74

	// ExitTempFail indicates temporary failure (user is invited to retry).
	ExitTempFail = 75

	// ExitProtocol indicates remote error in protocol.
	ExitProtocol = 76

	// ExitNoPerm indicates permission denied.
	ExitNoPerm = 77

	// ExitConfig indicates configuration error.
	ExitConfig = 78
)
