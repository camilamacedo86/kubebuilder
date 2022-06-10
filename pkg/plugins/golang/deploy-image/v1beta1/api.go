/*
Copyright 2022 The Kubernetes Authors.

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
	"errors"
	"fmt"
	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/deploy-image/v1beta1/scaffolds"
)

var _ plugin.CreateAPISubcommand = &createAPISubcommand{}

type createAPISubcommand struct {
	config config.Config
	resource *resource.Resource
	image string
}

func (p *createAPISubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {
	subcmdMeta.Description = `Scaffold the code implementation to deploy and manage your Operand which is represented
by the API (GVK) informed. Note that you must provide an image.

<TODO: here we will add the help of the plugin>
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Create a frigates API with Group: ship, Version: v1beta1 and Kind: Frigate
  %[1]s create api --group ship --version v1beta1 --kind Frigate

   <TODO here we will put the valid examples over how to call the plugin and its next steps>

  # Generate the manifests
  make manifests

  # Install CRDs into the Kubernetes cluster using kubectl apply
  make install

  # Regenerate code and run against the Kubernetes cluster configured by ~/.kube/config
  make run
`, cliMeta.CommandName)
}

func (p *createAPISubcommand) InjectConfig(c config.Config) error {
	p.config = c

	return nil
}

func (p *createAPISubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&p.image, "image", "", "inform the Operand image. <todo better text here>")
}

func (p *createAPISubcommand) InjectResource(res *resource.Resource) error {
	p.resource = res

	if len(p.image) == 0 {
		return fmt.Errorf("you must inform the image ")
	}

	// To scaffold the code implementation is required to have an API and controller
	//if !p.resource.HasAPI() || !p.resource.HasController() {
	//	return plugin.ExitError{
	//		Plugin: pluginName,
	//		Reason: "deploy-image/v1beta1 is only supported when API and controller are scaffolded",
	//	}
	//}

	return nil
}

func (p *createAPISubcommand) Scaffold(fs machinery.Filesystem) error {
	fmt.Println("updating scaffold with deploy-image/v1beta1 plugin...")

	scaffolder := scaffolds.NewAPIScaffolder(p.config, *p.resource, p.image)
	scaffolder.InjectFS(fs)
	err := scaffolder.Scaffold()
	if err != nil {
		return err
	}

	// Track the resources following a declarative approach
	cfg := pluginConfig{}
	if err := p.config.DecodePluginConfig(pluginKey, &cfg); errors.As(err, &config.UnsupportedFieldError{}) {
		// Config doesn't support per-plugin configuration, so we can't track them
	} else {
		// Fail unless they key wasn't found, which just means it is the first resource tracked
		if err != nil && !errors.As(err, &config.PluginKeyNotFoundError{}) {
			return err
		}

		cfg.Resources = append(cfg.Resources, p.resource.GVK)
		if err := p.config.EncodePluginConfig(pluginKey, cfg); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return nil
}
