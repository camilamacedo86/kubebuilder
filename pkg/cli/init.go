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

package cli

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	internalconfig "sigs.k8s.io/kubebuilder/v3/pkg/cli/internal/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
)

func (c *cli) newInitCmd() *cobra.Command {
	ctx := c.newInitContext()
	cmd := &cobra.Command{
		Use:     "init",
		Short:   "Initialize a new project",
		Long:    ctx.Description,
		Example: ctx.Examples,
		Run:     func(cmd *cobra.Command, args []string) {},
	}

	// Register --project-version on the dynamically created command
	// so that it shows up in help and does not cause a parse error.
	cmd.Flags().String(projectVersionFlag, c.defaultProjectVersion,
		fmt.Sprintf("project version, possible values: (%s)", strings.Join(c.getAvailableProjectVersions(), ", ")))
	// The --plugins flag can only be called to init projects v2+.
	if c.projectVersion != config.Version2 {
		cmd.Flags().StringSlice(pluginsFlag, nil,
			"Name and optionally version of the plugin to initialize the project with. "+
				fmt.Sprintf("Available plugins: (%s)", strings.Join(c.getAvailablePlugins(), ", ")))
	}

	// If only the help flag was set, return the command as is.
	if c.doGenericHelp {
		return cmd
	}

	// Lookup the plugin for projectVersion and bind it to the command.
	c.bindInit(ctx, cmd)
	return cmd
}

func (c cli) newInitContext() plugin.Context {
	return plugin.Context{
		CommandName: c.commandName,
		Description: `Initialize a new project.

For further help about a specific project version, set --project-version.
`,
		Examples: c.getInitHelpExamples(),
	}
}

func (c cli) getInitHelpExamples() string {
	var sb strings.Builder
	for _, version := range c.getAvailableProjectVersions() {
		rendered := fmt.Sprintf(`  # Help for initializing a project with version %s
  %s init --project-version=%s -h

`,
			version, c.commandName, version)
		sb.WriteString(rendered)
	}
	return strings.TrimSuffix(sb.String(), "\n\n")
}

func (c cli) getAvailableProjectVersions() (projectVersions []string) {
	versionSet := make(map[string]struct{})
	for version, versionedPlugins := range c.pluginsFromOptions {
		for _, p := range versionedPlugins {
			// If there's at least one non-deprecated plugin per version, that
			// version is "available".
			if _, isDeprecated := p.(plugin.Deprecated); !isDeprecated {
				versionSet[version] = struct{}{}
				break
			}
		}
	}
	for version := range versionSet {
		projectVersions = append(projectVersions, strconv.Quote(version))
	}
	sort.Strings(projectVersions)
	return projectVersions
}

func (c cli) getAvailablePlugins() (pluginKeys []string) {
	keySet := make(map[string]struct{})
	for _, versionedPlugins := range c.pluginsFromOptions {
		for _, p := range versionedPlugins {
			// Only return non-deprecated plugins.
			if _, isDeprecated := p.(plugin.Deprecated); !isDeprecated {
				keySet[plugin.KeyFor(p)] = struct{}{}
			}
		}
	}
	for key := range keySet {
		pluginKeys = append(pluginKeys, strconv.Quote(key))
	}
	sort.Strings(pluginKeys)
	return pluginKeys
}

func (c cli) bindInit(ctx plugin.Context, cmd *cobra.Command) {
	var initPlugin plugin.Init
	for _, p := range c.resolvedPlugins {
		tmpPlugin, isValid := p.(plugin.Init)
		if isValid {
			if initPlugin != nil {
				err := fmt.Errorf("duplicate initialization plugins (%s, %s), use a more specific plugin key",
					plugin.KeyFor(initPlugin), plugin.KeyFor(p))
				cmdErrNoHelp(cmd, err)
				return
			}
			initPlugin = tmpPlugin
		}
	}

	if initPlugin == nil {
		err := fmt.Errorf("relevant plugins do not provide an initialization plugin")
		cmdErrNoHelp(cmd, err)
		return
	}

	cfg := internalconfig.New(internalconfig.DefaultPath)
	cfg.Version = c.projectVersion

	subcommand := initPlugin.GetInitSubcommand()
	subcommand.InjectConfig(&cfg.Config)
	subcommand.BindFlags(cmd.Flags())
	subcommand.UpdateContext(&ctx)
	cmd.Long = ctx.Description
	cmd.Example = ctx.Examples
	cmd.RunE = func(*cobra.Command, []string) error {
		// Check if a config is initialized in the command runner so the check
		// doesn't erroneously fail other commands used in initialized projects.
		_, err := internalconfig.Read()
		if err == nil || os.IsExist(err) {
			log.Fatal("config already initialized")
		}
		if err := subcommand.Run(); err != nil {
			return fmt.Errorf("failed to initialize project with %q: %v", plugin.KeyFor(initPlugin), err)
		}
		return cfg.Save()
	}
}
