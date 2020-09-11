// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ozonep/drone-runner-kube/pkg/client"
	"github.com/ozonep/drone-runner-kube/pkg/environ"
	"github.com/ozonep/drone-runner-kube/pkg/logger"
	"github.com/ozonep/drone-runner-kube/pkg/manifest"
	"github.com/ozonep/drone-runner-kube/pkg/pipeline"
	"github.com/ozonep/drone-runner-kube/pkg/secret"

	"github.com/ozonep/drone/pkg/drone"
	"github.com/ozonep/drone/pkg/envsubst"
)

var noContext = context.Background()

// Runner runs the pipeline.
type Runner struct {
	// Machine provides the runner with the name of the host
	// machine executing the pipeline.
	Machine string

	// Environ defines global environment variables that can
	// be interpolated into the yaml using bash substitution.
	Environ map[string]string

	// Client is the remote client responsible for interacting
	// with the central server.
	Client client.Client

	// Compiler is responsible for compiling the pipeline
	// configuration to the intermediate representation.
	Compiler Compiler

	// Reporter reports pipeline status and logs back to the
	// remote server.
	Reporter pipeline.Reporter

	// Execer is responsible for executing intermediate
	// representation of the pipeline and returns its results.
	Exec func(context.Context, Spec, *pipeline.State) error

	// Lint is responsible for linting the pipeline
	// and failing if any rules are broken.
	Lint func(manifest.Resource, *drone.Repo) error

	// Match is an optional function that returns true if the
	// repository or build match user-defined criteria. This is
	// intended as a security measure to prevent a runner from
	// processing an unwanted pipeline.
	Match func(*drone.Repo, *drone.Build) bool

	// Lookup is a helper function that extracts the resource
	// from the manifest by name.
	Lookup func(string, *manifest.Manifest) (manifest.Resource, error)
}

// Run runs the pipeline stage.
func (s *Runner) Run(ctx context.Context, stage *drone.Stage) error {
	log := logger.FromContext(ctx).
		WithField("stage.id", stage.ID).
		WithField("stage.name", stage.Name).
		WithField("stage.number", stage.Number)

	log.Debug("stage received")

	// delivery to a single agent is not guaranteed, which means
	// we need confirm receipt. The first agent that confirms
	// receipt of the stage can assume ownership.

	stage.Machine = s.Machine
	err := s.Client.Accept(ctx, stage)
	if err != nil && err == client.ErrOptimisticLock {
		log.Debug("stage accepted by another runner")
		return nil
	}
	if err != nil {
		log.WithError(err).Error("cannot accept stage")
		return err
	}

	log.Debug("stage accepted")

	data, err := s.Client.Detail(ctx, stage)
	if err != nil {
		log.WithError(err).Error("cannot get stage details")
		return err
	}

	log = log.WithField("repo.id", data.Repo.ID).
		WithField("repo.namespace", data.Repo.Namespace).
		WithField("repo.name", data.Repo.Name).
		WithField("build.id", data.Build.ID).
		WithField("build.number", data.Build.Number)

	log.Debug("stage details fetched")
	return s.run(ctx, stage, data)
}

// RunAccepted runs a pipeline stage that has already been
// accepted and assigned to a runner.
func (s *Runner) RunAccepted(ctx context.Context, id int64) error {
	log := logger.FromContext(ctx).WithField("stage.id", id)
	log.Debug("stage received")

	data, err := s.Client.Detail(ctx, &drone.Stage{ID: id})
	if err != nil {
		log.WithError(err).Error("cannot get stage details")
		return err
	}

	log = log.WithField("repo.id", data.Repo.ID).
		WithField("repo.namespace", data.Repo.Namespace).
		WithField("repo.name", data.Repo.Name).
		WithField("build.id", data.Build.ID).
		WithField("build.number", data.Build.Number)

	log.Debug("stage details fetched")
	return s.run(ctx, data.Stage, data)
}

func (s *Runner) run(ctx context.Context, stage *drone.Stage, data *client.Context) error {
	log := logger.FromContext(ctx).
		WithField("repo.id", data.Repo.ID).
		WithField("stage.id", stage.ID).
		WithField("stage.name", stage.Name).
		WithField("stage.number", stage.Number).
		WithField("repo.namespace", data.Repo.Namespace).
		WithField("repo.name", data.Repo.Name).
		WithField("build.id", data.Build.ID).
		WithField("build.number", data.Build.Number)

	ctxdone, cancel := context.WithCancel(ctx)
	defer cancel()

	timeout := time.Duration(data.Repo.Timeout) * time.Minute
	ctxtimeout, cancel := context.WithTimeout(ctxdone, timeout)
	defer cancel()

	ctxcancel, cancel := context.WithCancel(ctxtimeout)
	defer cancel()

	// next we opens a connection to the server to watch for
	// cancellation requests. If a build is cancelled the running
	// stage should also be cancelled.
	go func() {
		done, _ := s.Client.Watch(ctxdone, data.Build.ID)
		if done {
			cancel()
			log.Debugln("received cancellation")
		} else {
			log.Debugln("done listening for cancellations")
		}
	}()

	envs := environ.Combine(
		s.Environ,
		environ.System(data.System),
		environ.Repo(data.Repo),
		environ.Build(data.Build),
		environ.Stage(stage),
		environ.Link(data.Repo, data.Build, data.System),
		data.Build.Params,
	)

	// string substitution function ensures that string
	// replacement variables are escaped and quoted if they
	// contain a newline character.
	subf := func(k string) string {
		v := envs[k]
		if strings.Contains(v, "\n") {
			v = fmt.Sprintf("%q", v)
		}
		return v
	}

	state := &pipeline.State{
		Build:  data.Build,
		Stage:  stage,
		Repo:   data.Repo,
		System: data.System,
	}

	// evaluates whether or not the agent can process the
	// pipeline. An agent may choose to reject a repository
	// or build for security reasons.
	if s.Match != nil && s.Match(data.Repo, data.Build) == false {
		log.Error("cannot process stage, access denied")
		state.FailAll(errors.New("insufficient permission to run the pipeline"))
		return s.Reporter.ReportStage(noContext, state)
	}

	// evaluates string replacement expressions and returns an
	// update configuration file string.
	config, err := envsubst.Eval(string(data.Config.Data), subf)
	if err != nil {
		log.WithError(err).Error("cannot emulate bash substitution")
		state.FailAll(err)
		return s.Reporter.ReportStage(noContext, state)
	}

	// parse the yaml configuration file.
	manifest, err := manifest.ParseString(config)
	if err != nil {
		log.WithError(err).Error("cannot parse configuration file")
		state.FailAll(err)
		return s.Reporter.ReportStage(noContext, state)
	}

	// find the named stage in the yaml configuration file.
	resource, err := s.Lookup(stage.Name, manifest)
	if err != nil {
		log.WithError(err).Error("cannot find pipeline resource")
		state.FailAll(err)
		return s.Reporter.ReportStage(noContext, state)
	}

	// lint the pipeline configuration and fail the build
	// if any linting rules are broken.
	err = s.Lint(resource, data.Repo)
	if err != nil {
		log.WithError(err).Error("cannot accept configuration")
		state.FailAll(err)
		return s.Reporter.ReportStage(noContext, state)
	}

	secrets := secret.Combine(
		secret.Static(data.Secrets),
		secret.Encrypted(),
	)

	// compile the yaml configuration file to an intermediate
	// representation, and then
	args := CompilerArgs{
		Pipeline: resource,
		Manifest: manifest,
		Build:    data.Build,
		Stage:    stage,
		Repo:     data.Repo,
		System:   data.System,
		Netrc:    data.Netrc,
		Secret:   secrets,
	}

	spec := s.Compiler.Compile(ctx, args)
	for i := 0; i < spec.StepLen(); i++ {
		src := spec.StepAt(i)

		// steps that are skipped are ignored and are not stored
		// in the drone database, nor displayed in the UI.
		if src.GetRunPolicy() == RunNever {
			continue
		}
		stage.Steps = append(stage.Steps, &drone.Step{
			Name:      src.GetName(),
			Number:    len(stage.Steps) + 1,
			StageID:   stage.ID,
			Status:    drone.StatusPending,
			ErrIgnore: src.GetErrPolicy() == ErrIgnore,
		})
	}

	stage.Started = time.Now().Unix()
	stage.Status = drone.StatusRunning
	if err := s.Client.Update(ctx, stage); err != nil {
		log.WithError(err).Error("cannot update stage")
		return err
	}

	log.Debug("updated stage to running")

	ctxcancel = logger.WithContext(ctxcancel, log)
	err = s.Exec(ctxcancel, spec, state)
	if err != nil {
		log.WithError(err).
			WithField("duration", stage.Stopped-stage.Started).
			Debug("stage failed")
		return err
	}
	log.WithField("duration", stage.Stopped-stage.Started).
		Debug("updated stage to complete")
	return nil
}
