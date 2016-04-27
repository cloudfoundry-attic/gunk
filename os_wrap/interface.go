package os_wrap

import real_os "os"

//go:generate counterfeiter -o osfakes/fake_os.go . Os

type Os interface {
	Open(name string) (*real_os.File, error)
}
