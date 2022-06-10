/*
Copyright 2020 The Kubernetes Authors.
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

package scaffolds

import (
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/deploy-image/v1beta1/scaffolds/internal/templates"
)

var _ plugins.Scaffolder = &createScaffolder{}

type createScaffolder struct {
	config config.Config
	image  string

	// fs is the filesystem that will be used by the scaffolder
	fs machinery.Filesystem
}

// NewCreateScaffolder returns a new Scaffolder for configuration create operations
func NewCreateScaffolder(config config.Config, image string) plugins.Scaffolder {
	return &createScaffolder{
		config: config,
		image:  image,
	}
}

// InjectFS implements cmdutil.Scaffolder
func (s *createScaffolder) InjectFS(fs machinery.Filesystem) {
	s.fs = fs
}

func (s *createScaffolder) Scaffold() error {
	scaffolder := NewCreateScaffolder(s.config, s.image)

	// Initialize the machinery.Scaffold that will write the files to disk
	scaffold := machinery.NewScaffold(s.fs,
		machinery.WithImage(s.image),
		machinery.WithConfig(s.config),
	)

	err := scaffold.Execute(
		&templates.Controller{Image: s.image},
	)
	if err != nil {
		return fmt.Errorf("error updating scaffold: %w", err)
	}
	return scaffolder.Scaffold()
}
