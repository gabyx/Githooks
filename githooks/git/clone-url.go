package git

import (
	"regexp"
	"strings"

	strs "github.com/gabyx/githooks/githooks/strings"
)

var reURLScheme *regexp.Regexp = regexp.MustCompile(`(?m)^[^:/?#]+://`)

var reShortSCPSyntax = regexp.MustCompile(
	`(?m)^(?:(?P<user>.+)@)?(?P<host>.+[^:]):(?P<path>[^:].*)`,
)
var reRemoteHelperSyntax = regexp.MustCompile(`(?m)^(?P<transport>.+)::(?P<address>.+)`)

// IsCloneURLALocalPath checks if the clone url is a local path.
// Thats the case if its not a URL Scheme,
// not a short SCP syntax and not
// a remote transport helper syntax.
// The problem arises on Windows with drive letters, since `C:/a/b`
// can technically be a short SCP syntax, we require at
// least 2 letters for the host name.
func IsCloneURLALocalPath(url string) bool {
	return !IsCloneURLANormalURL(url) &&
		!reShortSCPSyntax.MatchString(url) &&
		!reRemoteHelperSyntax.MatchString(url)
}

// IsCloneURLANormalURL checks if `url` is a normal url.
// Containing `<scheme>://` at the beginning.
func IsCloneURLANormalURL(url string) bool {
	return reURLScheme.MatchString(url)
}

// ShortSCP represents a short SCP
// syntax and corresponds to regex `reShortSCPSyntax`.
type ShortSCP []string

// ParseSCPSyntax parses the url as a short SCP syntax and reporting
// the user, host and path if not nil.
func ParseSCPSyntax(url string) ShortSCP {
	if m := reShortSCPSyntax.FindStringSubmatch(url); m != nil {
		return m[1:]
	}

	return nil
}

// String returns the whole short scp syntax as string.
func (scp ShortSCP) String() string {
	if strs.IsEmpty(scp[0]) {
		return scp[1] + ":" + scp[2]
	}

	return scp[0] + "@" + scp[1] + ":" + scp[2]
}

// IsCloneURLARemoteHelperSyntax checks if `url` is a remote helper syntax.
// https://git-scm.com/docs/gitremote-helpers
func IsCloneURLARemoteHelperSyntax(url string) bool {
	return reRemoteHelperSyntax.MatchString(url)
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
	return strings.HasPrefix(url, "file://")
}
