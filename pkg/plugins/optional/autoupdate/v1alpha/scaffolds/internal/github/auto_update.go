/*
Copyright 2025 The Kubernetes Authors.

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

package github

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
)

var _ machinery.Template = &AutoUpdate{}

// AutoUpdate scaffolds the GitHub Action to lint the project
type AutoUpdate struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin

	// UseGHModels indicates whether to enable GitHub Models AI summary
	UseGHModels bool
	// NotifyOnly when true scaffolds a notify-only workflow (open Issue only, no PR or branch push)
	NotifyOnly bool
}

// SetTemplateDefaults implements machinery.Template
func (f *AutoUpdate) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join(".github", "workflows", "auto_update.yml")
	}

	f.TemplateBody = autoUpdateTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const autoUpdateTemplate = `name: Auto Update

# {{ if .NotifyOnly }}Notify-only mode: workflow runs the update in CI and opens an Issue (no branch push or PR).
# Only issues: write{{ if .UseGHModels }} and models: read{{ end }} are required.{{ else }}
# The 'kubebuilder alpha update' command requires write access to the repository to create a branch
# with the update files and allow you to open a pull request using the link provided in the issue.
# The branch is named
# kubebuilder-update-from-<from-version>-to-<to-version> by default. To protect your codebase, ensure branch
# protection rules are configured for your main branches so updates cannot be merged without review.{{ end }}
permissions:
{{ if .NotifyOnly }}
  issues: write{{ if .UseGHModels }}
  models: read{{ end }}
{{ else }}
  contents: write       # Create and push the update branch
  pull-requests: write  # Create the Pull Request
  issues: write        # Create the Issue (default: open both PR and Issue){{ if .UseGHModels }}
  models: read         # AI summary for PR description and Issue comment{{ end }}
{{ end }}

on:
  workflow_dispatch:
  schedule:
    - cron: "0 0 * * 2" # Every Tuesday at 00:00 UTC

jobs:
  auto-update:
    runs-on: ubuntu-latest
    env:
      GH_TOKEN: {{ "${{ secrets.GITHUB_TOKEN }}" }}

    # Checkout the repository.
    steps:
    - name: Checkout repository
      uses: actions/checkout@v4
      with:
        token: {{ "${{ secrets.GITHUB_TOKEN }}" }}
        fetch-depth: 0

    # Configure Git to create commits with the GitHub Actions bot.
    - name: Configure Git
      run: |
        git config --global user.name "github-actions[bot]"
        git config --global user.email "github-actions[bot]@users.noreply.github.com"

    # Set up Go environment.
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: stable

    # Install Kubebuilder.
    - name: Install Kubebuilder
      run: |
        curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
        chmod +x kubebuilder
        sudo mv kubebuilder /usr/local/bin/
        kubebuilder version
{{ if .UseGHModels }}
    # Install Models extension for GitHub CLI.
    - name: Install gh-models extension
      run: |
        gh extension install github/gh-models --force
        gh models --help >/dev/null
{{ end }}
    # Run the Kubebuilder alpha update command.
    # More info: https://kubebuilder.io/reference/commands/alpha_update
    - name: Run kubebuilder alpha update
{{ if .NotifyOnly }}
      # Notify-only: no --push, no --open-gh-pr. Opens an Issue recommending to run the update locally.
      # --force: Completes the merge even if conflicts occur (run happens in CI; branch is not pushed).
      # --restore-path: Preserves specified paths when squashing.
      # --open-gh-issue: Creates an Issue notifying that a new release is available.{{ if .UseGHModels }}
      # --use-gh-models: Adds an AI-generated comment to the Issue with change overview and conflict guidance.{{ end }}
      run: |
        kubebuilder alpha update \
          --force \
          --restore-path .github/workflows \
          --open-gh-issue{{ if .UseGHModels }} \
          --use-gh-models{{ end }}
{{ else }}
      # Executes the update command with specified flags.
      # --force: Completes the merge even if conflicts occur, leaving conflict markers.
      # --push: Automatically pushes the resulting output branch to the 'origin' remote.
      # --restore-path: Preserves specified paths (e.g., CI workflow files) when squashing.
      # --open-gh-issue: Creates a GitHub Issue with a link for opening a PR for review.{{ if .UseGHModels }}
      # --use-gh-models: AI summary is used for the PR description and as an Issue comment.{{ else }}
      #
      # WARNING: This workflow does not use GitHub Models AI by default.
      # To enable AI usage, you need permissions to use GitHub Models.
      # If you have the required permissions, re-run:
      #   kubebuilder edit --plugins="autoupdate/v1-alpha" --use-gh-models
{{ end }}
      run: |
        kubebuilder alpha update \
          --force \
          --push \
          --restore-path .github/workflows \
          --open-gh-pr \
          --open-gh-issue{{ if .UseGHModels }} \
          --use-gh-models{{ end }}
{{ end }}
`
