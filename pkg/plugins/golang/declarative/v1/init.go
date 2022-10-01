/*
Copyright 2021 The Kubernetes Authors.

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

package v1

import (
	"fmt"
	"path/filepath"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin/util"
)

var _ plugin.InitSubcommand = &initSubcommand{}

type initSubcommand struct {
	config config.Config
}

func (p *initSubcommand) InjectConfig(c config.Config) error {
	p.config = c

	return nil
}

func (p *initSubcommand) Scaffold(fs machinery.Filesystem) error {
	return updateDockerfile()
}

// updateDockerfile will add channels staging required for declarative plugin
func updateDockerfile() error {
	fmt.Println("updating Dockerfile to add channels/ directory in the image")
	dokerfile := filepath.Join("Dockerfile")

	isLegacyLayout, err := util.HasFragment(dokerfile, "COPY controllers/ controllers/")
	if err != nil {
		return err
	}
	if isLegacyLayout {
		// nolint:lll
		err := util.InsertCode(dokerfile, "COPY controllers/ controllers/",
			"\n# https://github.com/kubernetes-sigs/kubebuilder-declarative-pattern/blob/master/docs/addon/walkthrough/README.md#adding-a-manifest\n# Stage channels and make readable\nCOPY channels/ /channels/\nRUN chmod -R a+rx /channels/")
		if err != nil {
			return err
		}
	} else {
		// nolint:lll
		err := util.InsertCode(dokerfile,
			"COPY pkg/controllers/ pkg/controllers/",
			"\n# https://github.com/kubernetes-sigs/kubebuilder-declarative-pattern/blob/master/docs/addon/walkthrough/README.md#adding-a-manifest\n# Stage channels and make readable\nCOPY channels/ /channels/\nRUN chmod -R a+rx /channels/")
		if err != nil {
			return err
		}
	}

	return nil
}
