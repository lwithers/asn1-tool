package util

// UsageError should be returned by a command action if there is a problem with
// the arguments. The error within will be reported to the user, and help for
// the given action will be displayed.
type UsageError struct {
	Problem string
}

func (u *UsageError) Error() string {
	return u.Problem
}
