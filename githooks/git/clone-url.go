package git

import "regexp"

var reURLScheme *regexp.Regexp = regexp.MustCompile(`(?m)^[^:/?#]+://`)
var reShortSCPSyntax = regexp.MustCompile(`(?m)^(?m)(?P<user>.+@)?(?P<host>.*[^:]):(?P<path>[^:].*)`)
var reRemoteHelperSyntax = regexp.MustCompile(`(?m)^(?P<transport>.+)::(?P<address>.+)`)
var reFileURLScheme = regexp.MustCompile(`(?m)^file://`)

// IsCloneURLALocalPath checks if the clone url is local path.
// Thats the case if its not a URL Scheme or a short SCP syntax
// or a remote transport helper syntax.
func IsCloneURLALocalPath(url string) bool {
	return !reURLScheme.MatchString(url) &&
		!reShortSCPSyntax.MatchString(url) &&
		!reRemoteHelperSyntax.MatchString(url)
}

// ParseSCPSyntax parses the url as a short SCP syntax and reporting
// the user, host and path if not nil.
func ParseSCPSyntax(url string) []string {
	if m := reShortSCPSyntax.FindStringSubmatch(url); m != nil {
		return m[1:]
	}

	return nil
}

// ParseRemoteHelperSyntax parses the url as a remote helper syntax and reporting
// the transport  and address string if not nil.
// https://git-scm.com/docs/gitremote-helpers
func ParseRemoteHelperSyntax(url string) []string {
	if m := reRemoteHelperSyntax.FindStringSubmatch(url); m != nil {
		return m[1:]
	}

	return nil
}

// IsCloneURLALocalURL checks if the clone url is a url to a local directory.
// Thats the case only for `file://`.
func IsCloneURLALocalURL(url string) bool {
	return reFileURLScheme.MatchString(url)
}
