// Package hook should be imported for side effect if you want configcache.AddHook to be called via init.
package hook

import configcache "github.com/nguyengg/go-aws-commons/metrics/config-cache"

func init() {
	configcache.AddHook()
}
