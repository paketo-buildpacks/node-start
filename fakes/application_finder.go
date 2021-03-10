package fakes

import "sync"

type ApplicationFinder struct {
	FindCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			WorkingDir string
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(string) (string, error)
	}
}

func (f *ApplicationFinder) Find(param1 string) (string, error) {
	f.FindCall.Lock()
	defer f.FindCall.Unlock()
	f.FindCall.CallCount++
	f.FindCall.Receives.WorkingDir = param1
	if f.FindCall.Stub != nil {
		return f.FindCall.Stub(param1)
	}
	return f.FindCall.Returns.String, f.FindCall.Returns.Error
}
