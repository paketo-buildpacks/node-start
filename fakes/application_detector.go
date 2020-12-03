package fakes

import "sync"

type ApplicationDetector struct {
	DetectCall struct {
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

func (f *ApplicationDetector) Detect(param1 string) (string, error) {
	f.DetectCall.Lock()
	defer f.DetectCall.Unlock()
	f.DetectCall.CallCount++
	f.DetectCall.Receives.WorkingDir = param1
	if f.DetectCall.Stub != nil {
		return f.DetectCall.Stub(param1)
	}
	return f.DetectCall.Returns.String, f.DetectCall.Returns.Error
}
