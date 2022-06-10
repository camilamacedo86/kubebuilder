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

package v1beta1

import (
	"fmt"

	"github.com/spf13/pflag"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/deploy-image/v1beta1/scaffolds"
)

var _ plugin.CreateSubcommand = &createSubcommand{}

type createSubcommand struct {
	config config.Config
	image  string
}

func (p *createSubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {
	subcmdMeta.Description = `This command will create the project for deploying an image.
`
	subcmdMeta.Examples = fmt.Sprintf(`	# Enable the image layout
  %[1]s create api --kind Guest --group webapp --versionv1 --image=imagename
`, cliMeta.CommandName)
}

func (p *createSubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&p.image, "image", "", "inform the Operand image.")
}

func (p *createSubcommand) InjectConfig(c config.Config) error {
	p.config = c

	return nil
}

func (p *createSubcommand) Scaffold(fs machinery.Filesystem) error {
	scaffolder := scaffolds.NewCreateScaffolder(p.config, p.image)
	scaffolder.InjectFS(fs)
	return scaffolder.Scaffold()
}
