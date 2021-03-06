// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"context"
	"sync"

	"github.com/bborbe/hue/pkg/check"
)

type Check struct {
	ApplyStub        func(context.Context) error
	applyMutex       sync.RWMutex
	applyArgsForCall []struct {
		arg1 context.Context
	}
	applyReturns struct {
		result1 error
	}
	applyReturnsOnCall map[int]struct {
		result1 error
	}
	NameStub        func() string
	nameMutex       sync.RWMutex
	nameArgsForCall []struct {
	}
	nameReturns struct {
		result1 string
	}
	nameReturnsOnCall map[int]struct {
		result1 string
	}
	SatisfiedStub        func(context.Context) (bool, error)
	satisfiedMutex       sync.RWMutex
	satisfiedArgsForCall []struct {
		arg1 context.Context
	}
	satisfiedReturns struct {
		result1 bool
		result2 error
	}
	satisfiedReturnsOnCall map[int]struct {
		result1 bool
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *Check) Apply(arg1 context.Context) error {
	fake.applyMutex.Lock()
	ret, specificReturn := fake.applyReturnsOnCall[len(fake.applyArgsForCall)]
	fake.applyArgsForCall = append(fake.applyArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.ApplyStub
	fakeReturns := fake.applyReturns
	fake.recordInvocation("Apply", []interface{}{arg1})
	fake.applyMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *Check) ApplyCallCount() int {
	fake.applyMutex.RLock()
	defer fake.applyMutex.RUnlock()
	return len(fake.applyArgsForCall)
}

func (fake *Check) ApplyCalls(stub func(context.Context) error) {
	fake.applyMutex.Lock()
	defer fake.applyMutex.Unlock()
	fake.ApplyStub = stub
}

func (fake *Check) ApplyArgsForCall(i int) context.Context {
	fake.applyMutex.RLock()
	defer fake.applyMutex.RUnlock()
	argsForCall := fake.applyArgsForCall[i]
	return argsForCall.arg1
}

func (fake *Check) ApplyReturns(result1 error) {
	fake.applyMutex.Lock()
	defer fake.applyMutex.Unlock()
	fake.ApplyStub = nil
	fake.applyReturns = struct {
		result1 error
	}{result1}
}

func (fake *Check) ApplyReturnsOnCall(i int, result1 error) {
	fake.applyMutex.Lock()
	defer fake.applyMutex.Unlock()
	fake.ApplyStub = nil
	if fake.applyReturnsOnCall == nil {
		fake.applyReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.applyReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *Check) Name() string {
	fake.nameMutex.Lock()
	ret, specificReturn := fake.nameReturnsOnCall[len(fake.nameArgsForCall)]
	fake.nameArgsForCall = append(fake.nameArgsForCall, struct {
	}{})
	stub := fake.NameStub
	fakeReturns := fake.nameReturns
	fake.recordInvocation("Name", []interface{}{})
	fake.nameMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *Check) NameCallCount() int {
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	return len(fake.nameArgsForCall)
}

func (fake *Check) NameCalls(stub func() string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = stub
}

func (fake *Check) NameReturns(result1 string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = nil
	fake.nameReturns = struct {
		result1 string
	}{result1}
}

func (fake *Check) NameReturnsOnCall(i int, result1 string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = nil
	if fake.nameReturnsOnCall == nil {
		fake.nameReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.nameReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *Check) Satisfied(arg1 context.Context) (bool, error) {
	fake.satisfiedMutex.Lock()
	ret, specificReturn := fake.satisfiedReturnsOnCall[len(fake.satisfiedArgsForCall)]
	fake.satisfiedArgsForCall = append(fake.satisfiedArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.SatisfiedStub
	fakeReturns := fake.satisfiedReturns
	fake.recordInvocation("Satisfied", []interface{}{arg1})
	fake.satisfiedMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *Check) SatisfiedCallCount() int {
	fake.satisfiedMutex.RLock()
	defer fake.satisfiedMutex.RUnlock()
	return len(fake.satisfiedArgsForCall)
}

func (fake *Check) SatisfiedCalls(stub func(context.Context) (bool, error)) {
	fake.satisfiedMutex.Lock()
	defer fake.satisfiedMutex.Unlock()
	fake.SatisfiedStub = stub
}

func (fake *Check) SatisfiedArgsForCall(i int) context.Context {
	fake.satisfiedMutex.RLock()
	defer fake.satisfiedMutex.RUnlock()
	argsForCall := fake.satisfiedArgsForCall[i]
	return argsForCall.arg1
}

func (fake *Check) SatisfiedReturns(result1 bool, result2 error) {
	fake.satisfiedMutex.Lock()
	defer fake.satisfiedMutex.Unlock()
	fake.SatisfiedStub = nil
	fake.satisfiedReturns = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *Check) SatisfiedReturnsOnCall(i int, result1 bool, result2 error) {
	fake.satisfiedMutex.Lock()
	defer fake.satisfiedMutex.Unlock()
	fake.SatisfiedStub = nil
	if fake.satisfiedReturnsOnCall == nil {
		fake.satisfiedReturnsOnCall = make(map[int]struct {
			result1 bool
			result2 error
		})
	}
	fake.satisfiedReturnsOnCall[i] = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *Check) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.applyMutex.RLock()
	defer fake.applyMutex.RUnlock()
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	fake.satisfiedMutex.RLock()
	defer fake.satisfiedMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *Check) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ check.Check = new(Check)
