/*
Copyright 2019 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package local

import (
	"context"
	"io"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/jib"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (b *Builder) buildJibMaven(ctx context.Context, out io.Writer, workspace string, artifact *latest.JibMavenArtifact, tag string) (string, error) {
	if b.pushImages {
		return b.buildJibMavenToRegistry(ctx, out, workspace, artifact, tag)
	}
	return b.buildJibMavenToDocker(ctx, out, workspace, artifact, tag)
}

func (b *Builder) buildJibMavenToDocker(ctx context.Context, out io.Writer, workspace string, artifact *latest.JibMavenArtifact, tag string) (string, error) {
	// If this is a multi-module project, we require `package` be bound to jib:dockerBuild
	if artifact.Module != "" {
		if err := verifyJibPackageGoal(ctx, "dockerBuild", workspace, artifact); err != nil {
			return "", err
		}
	}

	args := jib.GenerateMavenArgs("dockerBuild", tag, artifact, b.skipTests)
	if err := b.runMavenCommand(ctx, out, workspace, args); err != nil {
		return "", err
	}

	return b.localDocker.ImageID(ctx, tag)
}

func (b *Builder) buildJibMavenToRegistry(ctx context.Context, out io.Writer, workspace string, artifact *latest.JibMavenArtifact, tag string) (string, error) {
	// If this is a multi-module project, we require `package` be bound to jib:build
	if artifact.Module != "" {
		if err := verifyJibPackageGoal(ctx, "build", workspace, artifact); err != nil {
			return "", err
		}
	}

	args := jib.GenerateMavenArgs("build", tag, artifact, b.skipTests)
	if err := b.runMavenCommand(ctx, out, workspace, args); err != nil {
		return "", err
	}

	return docker.RemoteDigest(tag, b.insecureRegistries)
}

// verifyJibPackageGoal verifies that the referenced module has `package` bound to a single jib goal.
// It returns `nil` if the goal is matched, and an error if there is a mismatch.
func verifyJibPackageGoal(ctx context.Context, requiredGoal string, workspace string, artifact *latest.JibMavenArtifact) error {
	// cannot use --non-recursive
	command := []string{"--quiet", "--projects", artifact.Module, "jib:_skaffold-package-goals"}
	if artifact.Profile != "" {
		command = append(command, "--activate-profiles", artifact.Profile)
	}

	cmd := jib.MavenCommand.CreateCommand(ctx, workspace, command)
	logrus.Debugf("Looking for jib bound package goals for %s: %s, %v", workspace, cmd.Path, cmd.Args)
	stdout, err := util.RunCmdOut(&cmd)
	if err != nil {
		return errors.Wrap(err, "could not obtain jib package goals")
	}
	goals := util.NonEmptyLines(stdout)
	logrus.Debugf("jib bound package goals for %s %s: %v (%d)", workspace, artifact.Module, goals, len(goals))
	if len(goals) != 1 {
		return errors.New("skaffold requires a single jib goal bound to 'package'")
	}
	if goals[0] != requiredGoal {
		return errors.Errorf("skaffold `push` setting requires 'package' be bound to 'jib:%s'", requiredGoal)
	}
	return nil
}

func (b *Builder) runMavenCommand(ctx context.Context, out io.Writer, workspace string, args []string) error {
	cmd := jib.MavenCommand.CreateCommand(ctx, workspace, args)
	cmd.Env = append(util.OSEnviron(), b.localDocker.ExtraEnv()...)
	cmd.Stdout = out
	cmd.Stderr = out

	logrus.Infof("Building %s: %s, %v", workspace, cmd.Path, cmd.Args)
	if err := util.RunCmd(&cmd); err != nil {
		return errors.Wrap(err, "maven build failed")
	}

	return nil
}
