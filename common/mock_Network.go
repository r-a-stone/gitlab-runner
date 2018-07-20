// Code generated by mockery v1.0.0

// This comment works around https://github.com/vektra/mockery/issues/155

package common

import io "io"
import mock "github.com/stretchr/testify/mock"

// MockNetwork is an autogenerated mock type for the Network type
type MockNetwork struct {
	mock.Mock
}

// DownloadArtifacts provides a mock function with given fields: config, artifactsFile
func (_m *MockNetwork) DownloadArtifacts(config JobCredentials, artifactsFile string) DownloadState {
	ret := _m.Called(config, artifactsFile)

	var r0 DownloadState
	if rf, ok := ret.Get(0).(func(JobCredentials, string) DownloadState); ok {
		r0 = rf(config, artifactsFile)
	} else {
		r0 = ret.Get(0).(DownloadState)
	}

	return r0
}

// PatchTrace provides a mock function with given fields: config, jobCredentials, tracePart
func (_m *MockNetwork) PatchTrace(config RunnerConfig, jobCredentials *JobCredentials, tracePart JobTracePatch) UpdateState {
	ret := _m.Called(config, jobCredentials, tracePart)

	var r0 UpdateState
	if rf, ok := ret.Get(0).(func(RunnerConfig, *JobCredentials, JobTracePatch) UpdateState); ok {
		r0 = rf(config, jobCredentials, tracePart)
	} else {
		r0 = ret.Get(0).(UpdateState)
	}

	return r0
}

// ProcessJob provides a mock function with given fields: config, buildCredentials
func (_m *MockNetwork) ProcessJob(config RunnerConfig, buildCredentials *JobCredentials) JobTrace {
	ret := _m.Called(config, buildCredentials)

	var r0 JobTrace
	if rf, ok := ret.Get(0).(func(RunnerConfig, *JobCredentials) JobTrace); ok {
		r0 = rf(config, buildCredentials)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(JobTrace)
		}
	}

	return r0
}

// RegisterRunner provides a mock function with given fields: config, parameters
func (_m *MockNetwork) RegisterRunner(config RunnerCredentials, parameters RegisterRunnerParameters) *RegisterRunnerResponse {
	ret := _m.Called(config, parameters)

	var r0 *RegisterRunnerResponse
	if rf, ok := ret.Get(0).(func(RunnerCredentials, RegisterRunnerParameters) *RegisterRunnerResponse); ok {
		r0 = rf(config, parameters)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*RegisterRunnerResponse)
		}
	}

	return r0
}

// RequestJob provides a mock function with given fields: config
func (_m *MockNetwork) RequestJob(config RunnerConfig) (*JobResponse, bool) {
	ret := _m.Called(config)

	var r0 *JobResponse
	if rf, ok := ret.Get(0).(func(RunnerConfig) *JobResponse); ok {
		r0 = rf(config)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*JobResponse)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(RunnerConfig) bool); ok {
		r1 = rf(config)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// UnregisterRunner provides a mock function with given fields: config
func (_m *MockNetwork) UnregisterRunner(config RunnerCredentials) bool {
	ret := _m.Called(config)

	var r0 bool
	if rf, ok := ret.Get(0).(func(RunnerCredentials) bool); ok {
		r0 = rf(config)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// UpdateJob provides a mock function with given fields: config, jobCredentials, jobInfo
func (_m *MockNetwork) UpdateJob(config RunnerConfig, jobCredentials *JobCredentials, jobInfo UpdateJobInfo) UpdateState {
	ret := _m.Called(config, jobCredentials, jobInfo)

	var r0 UpdateState
	if rf, ok := ret.Get(0).(func(RunnerConfig, *JobCredentials, UpdateJobInfo) UpdateState); ok {
		r0 = rf(config, jobCredentials, jobInfo)
	} else {
		r0 = ret.Get(0).(UpdateState)
	}

	return r0
}

// UploadRawArtifacts provides a mock function with given fields: config, reader, options
func (_m *MockNetwork) UploadRawArtifacts(config JobCredentials, reader io.Reader, options ArtifactsOptions) UploadState {
	ret := _m.Called(config, reader, options)

	var r0 UploadState
	if rf, ok := ret.Get(0).(func(JobCredentials, io.Reader, ArtifactsOptions) UploadState); ok {
		r0 = rf(config, reader, options)
	} else {
		r0 = ret.Get(0).(UploadState)
	}

	return r0
}

// VerifyRunner provides a mock function with given fields: config
func (_m *MockNetwork) VerifyRunner(config RunnerCredentials) bool {
	ret := _m.Called(config)

	var r0 bool
	if rf, ok := ret.Get(0).(func(RunnerCredentials) bool); ok {
		r0 = rf(config)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
