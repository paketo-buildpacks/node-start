package fakes

import "sync"

type ApplicationFinder struct {
	FindCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			WorkingDir  string
			Launchpoint string
			ProjectPath string
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(string, string, string) (string, error)
	}
}

func (f *ApplicationFinder) Find(param1 string, param2 string, param3 string) (string, error) {
	f.FindCall.mutex.Lock()
	defer f.FindCall.mutex.Unlock()
	f.FindCall.CallCount++
	f.FindCall.Receives.WorkingDir = param1
	f.FindCall.Receives.Launchpoint = param2
	f.FindCall.Receives.ProjectPath = param3
	if f.FindCall.Stub != nil {
		return f.FindCall.Stub(param1, param2, param3)
	}
	return f.FindCall.Returns.String, f.FindCall.Returns.Error
}
