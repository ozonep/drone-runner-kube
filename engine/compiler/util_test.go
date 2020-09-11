// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"testing"

	"github.com/ozonep/drone-runner-kube/engine"
	"github.com/ozonep/drone-runner-kube/engine/resource"
	"github.com/ozonep/drone-runner-kube/pkg/manifest"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Test_isRunAlways(t *testing.T) {
	step := new(resource.Step)
	if isRunAlways(step) {
		t.Errorf("Want always run false if empty when clause")
	}
	step.When.Status.Include = []string{"success"}
	if isRunAlways(step) {
		t.Errorf("Want always run false if when success")
	}
	step.When.Status.Include = []string{"failure"}
	if isRunAlways(step) {
		t.Errorf("Want always run false if when faiure")
	}
	step.When.Status.Include = []string{"success", "failure"}
	if !isRunAlways(step) {
		t.Errorf("Want always run true if when success, failure")
	}
}

func Test_isRunOnFailure(t *testing.T) {
	step := new(resource.Step)
	if isRunOnFailure(step) {
		t.Errorf("Want run on failure false if empty when clause")
	}
	step.When.Status.Include = []string{"success"}
	if isRunOnFailure(step) {
		t.Errorf("Want run on failure false if when success")
	}
	step.When.Status.Include = []string{"failure"}
	if !isRunOnFailure(step) {
		t.Errorf("Want run on failure true if when faiure")
	}
	step.When.Status.Include = []string{"success", "failure"}
	if !isRunOnFailure(step) {
		t.Errorf("Want run on failure true if when success, failure")
	}
}

func Test_isGraph(t *testing.T) {
	spec := new(engine.Spec)
	spec.Steps = []*engine.Step{
		{DependsOn: []string{}},
	}
	if isGraph(spec) {
		t.Errorf("Expect is graph false if deps not exist")
	}
	spec.Steps[0].DependsOn = []string{"clone"}
	if !isGraph(spec) {
		t.Errorf("Expect is graph true if deps exist")
	}
}

func Test_configureSerial(t *testing.T) {
	before := new(engine.Spec)
	before.Steps = []*engine.Step{
		{Name: "build"},
		{Name: "test"},
		{Name: "deploy"},
	}

	after := new(engine.Spec)
	after.Steps = []*engine.Step{
		{Name: "build"},
		{Name: "test", DependsOn: []string{"build"}},
		{Name: "deploy", DependsOn: []string{"test"}},
	}
	configureSerial(before)

	opts := cmpopts.IgnoreUnexported(engine.Spec{})
	if diff := cmp.Diff(before, after, opts); diff != "" {
		t.Errorf("Unexpected serial configuration")
		t.Log(diff)
	}
}

func Test_convertStaticEnv(t *testing.T) {
	vars := map[string]*manifest.Variable{
		"username": {Value: "octocat"},
		"password": {Secret: "password"},
	}
	envs := convertStaticEnv(vars)
	want := map[string]string{"username": "octocat"}
	if diff := cmp.Diff(envs, want); diff != "" {
		t.Errorf("Unexpected environment variable set")
		t.Log(diff)
	}
}

func Test_convertSecretEnv(t *testing.T) {
	vars := map[string]*manifest.Variable{
		"USERNAME": {Value: "octocat"},
		"PASSWORD": {Secret: "password"},
	}
	envs := convertSecretEnv(vars)
	want := []*engine.SecretVar{
		{
			Name: "password",
			Env:  "PASSWORD",
		},
	}
	if diff := cmp.Diff(envs, want); diff != "" {
		t.Errorf("Unexpected secret list")
		t.Log(diff)
	}
}

func Test_configureCloneDeps(t *testing.T) {
	before := new(engine.Spec)
	before.Steps = []*engine.Step{
		{Name: "clone"},
		{Name: "backend"},
		{Name: "frontend"},
		{Name: "deploy", DependsOn: []string{
			"backend", "frontend",
		}},
	}

	after := new(engine.Spec)
	after.Steps = []*engine.Step{
		{Name: "clone"},
		{Name: "backend", DependsOn: []string{"clone"}},
		{Name: "frontend", DependsOn: []string{"clone"}},
		{Name: "deploy", DependsOn: []string{
			"backend", "frontend",
		}},
	}
	configureCloneDeps(before)

	opts := cmpopts.IgnoreUnexported(engine.Spec{})
	if diff := cmp.Diff(before, after, opts); diff != "" {
		t.Errorf("Unexpected dependency adjustment")
		t.Log(diff)
	}
}

func Test_removeCloneDeps(t *testing.T) {
	before := new(engine.Spec)
	before.Steps = []*engine.Step{
		{Name: "backend", DependsOn: []string{"clone"}},
		{Name: "frontend", DependsOn: []string{"clone"}},
		{Name: "deploy", DependsOn: []string{
			"backend", "frontend",
		}},
	}

	after := new(engine.Spec)
	after.Steps = []*engine.Step{
		{Name: "backend", DependsOn: []string{}},
		{Name: "frontend", DependsOn: []string{}},
		{Name: "deploy", DependsOn: []string{
			"backend", "frontend",
		}},
	}
	removeCloneDeps(before)

	opts := cmpopts.IgnoreUnexported(engine.Spec{})
	if diff := cmp.Diff(before, after, opts); diff != "" {
		t.Errorf("Unexpected result after removing clone deps")
		t.Log(diff)
	}
}

func Test_removeCloneDeps_CloneEnabled(t *testing.T) {
	before := new(engine.Spec)
	before.Steps = []*engine.Step{
		{Name: "clone"},
		{Name: "test", DependsOn: []string{"clone"}},
	}

	after := new(engine.Spec)
	after.Steps = []*engine.Step{
		{Name: "clone"},
		{Name: "test", DependsOn: []string{"clone"}},
	}
	removeCloneDeps(before)

	opts := cmpopts.IgnoreUnexported(engine.Spec{})
	if diff := cmp.Diff(before, after, opts); diff != "" {
		t.Errorf("Expect clone dependencies not removed")
		t.Log(diff)
	}
}
