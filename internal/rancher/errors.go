package rancher

import "errors"

// ErrTokenInvalid indicates the Rancher API token is invalid or expired.
var ErrTokenInvalid = errors.New("rancher token invalid or expired")
