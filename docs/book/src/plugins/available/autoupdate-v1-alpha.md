# AutoUpdate (`autoupdate/v1-alpha`)

Keeping your Kubebuilder project up to date with the latest improvements shouldn’t be a chore.
With a small amount of setup, you can receive **automatic Pull Request** suggestions whenever a new
Kubebuilder release is available — keeping your project **maintained, secure, and aligned with ecosystem changes**.

This automation uses the [`kubebuilder alpha update`][alpha-update-command] command with a **3-way merge strategy** to
refresh your project scaffold. By default, the workflow creates the update branch, **opens a Pull Request** (using `--open-gh-pr`), and **opens an Issue** (using `--open-gh-issue`). If you want only to be notified when a new release is available, use the **`--notify-only`** flag: the workflow will then open an Issue only, with reduced permissions (no branch push or PR).

<aside class="warning">
<h3>Protect your branches</h3>

By default the workflow creates and pushes a branch `kubebuilder-update-from-<from-version>-to-<to-version>`, **opens a Pull Request**, and **opens an Issue**. The workflow requests **contents: write**, **pull-requests: write**, and **issues: write**.

To keep your codebase safe, use branch protection rules so that changes cannot be merged without proper review.

</aside>

## When to Use It


- When you want to reduce the burden of keeping the project updated and well-maintained.
- When you want guidance and help from AI to know what changes are needed to keep your project up to date and to solve conflicts (requires `--use-gh-models` flag and GitHub Models permissions).

## How to Use It

- If you want to add the `autoupdate` plugin to your project:

```shell
kubebuilder edit --plugins="autoupdate/v1-alpha"
```

- If you want to create a new project with the `autoupdate` plugin:

```shell
kubebuilder init --plugins=go/v4,autoupdate/v1-alpha
```

### Notify-only workflow (`--notify-only`)

If you want only to be notified when a new release is available (no Pull Request, no branch push), use the **`--notify-only`** flag. The scaffolded workflow will use only **`issues: write`** (and `models: read` if you also use `--use-gh-models`):

```shell
kubebuilder edit --plugins="autoupdate/v1-alpha" --notify-only
```

The workflow runs the update in CI and opens an Issue that recommends running `kubebuilder alpha update` locally. You can combine with `--use-gh-models` for an AI summary comment on the issue.

If you already have a default workflow and prefer to change it manually, edit `.github/workflows/auto_update.yml`: in `permissions` keep only `issues: write` (and `models: read` if you use `--use-gh-models`), and in the run command remove `--push` and `--open-gh-pr` and keep `--open-gh-issue`.

### Optional: GitHub Models AI Summary

By default, the workflow works without GitHub Models to avoid permission errors.
If you want AI-generated summaries in your update issues:

```shell
kubebuilder edit --plugins="autoupdate/v1-alpha" --use-gh-models
```

<aside class="note">
<h1>Permissions required to use GitHub Models in GitHub Actions</h1>

To use GitHub Models in your workflows, organization and repository administrators must grant this permission.

**If you have admin access:**

1. Go to **Settings → Code and automation → Models**
2. Enable GitHub Models for your repository

**Don't see the Models option?**

Your organization or enterprise may have disabled it. Contact your administrator:

- Organization admins: [Managing Models in your organization][manage-org-models]
- Enterprise admins: [Managing Models at enterprise scale][manage-models-at-scale]

</aside>

## Default permissions

The scaffolded workflow requests the following permissions via `GITHUB_TOKEN`:

| Permission             | Purpose |
|-------------------------|--------|
| **contents: write**     | Create and push the update branch |
| **pull-requests: write**| Create the Pull Request (default behavior) |
| **issues: write**       | Create the Issue (default: workflow opens both PR and Issue; notify-only mode uses this only) |
| **models: read**        | Only if you use `--use-gh-models` (see [Optional: GitHub Models](#optional-github-models-ai-summary)) |

## How It Works

The plugin scaffolds a GitHub Actions workflow that checks for new Kubebuilder releases every week. When an update is available, it:

1. Creates a new branch with the merged changes and pushes it
2. Opens a **Pull Request** from that branch to your base branch
3. Opens an **Issue** that notifies about the release and recommends reviewing the PR (default)

With **`--use-gh-models`**, the same AI summary is used for the **PR description** and as an **Issue comment**.

**Example:** By default the workflow opens both a **Pull Request** and an **Issue**. With **`--notify-only`**, the workflow opens only an **Issue** that recommends running the update locally:

<img width="638" height="482" alt="Example Issue" src="https://github.com/user-attachments/assets/589fd16b-7709-4cd5-b169-fd53d69790d4" />

**With GitHub Models enabled** (optional), you also get AI-generated summaries:

<img width="582" height="646" alt="AI Summary" src="https://github.com/user-attachments/assets/d460a5af-5ca4-4dd5-afb8-7330dd6de148" />

**Conflict help** (when needed):

<img width="600" height="188" alt="Conflicts" src="https://github.com/user-attachments/assets/2142887a-730c-499a-94df-c717f09ab600" />

## Customizing the Workflow

The generated workflow uses the `kubebuilder alpha update` command with default flags. You can customize the workflow by editing `.github/workflows/auto_update.yml` to add additional flags:

**Default flags used:**
- `--force` - Continue even if conflicts occur (automation-friendly)
- `--push` - Push the output branch to remote (required for `--open-gh-pr`)
- `--restore-path .github/workflows` - Preserve CI workflows from base branch
- `--open-gh-pr` - Create a GitHub Pull Request from the update branch (default)
- `--open-gh-issue` - Create an Issue (default: both PR and Issue are opened)
- `--use-gh-models` - (optional) Use the same AI-generated summary for the **PR description** and as an **Issue comment** (overview of changes and conflict guidance)

**Additional available flags:**
- `--merge-message` - Custom commit message for clean merges
- `--conflict-message` - Custom commit message when conflicts occur
- `--from-version` - Specify the version to upgrade from
- `--to-version` - Specify the version to upgrade to
- `--output-branch` - Custom output branch name
- `--show-commits` - Keep full history instead of squashing
- `--git-config` - Pass per-invocation Git config

For complete documentation on all available flags, see the [`kubebuilder alpha update`][alpha-update-command] reference.

**Example: Customize commit messages**

Edit `.github/workflows/auto_update.yml`:

```yaml
- name: Run kubebuilder alpha update
  run: |
    kubebuilder alpha update \
      --force \
      --push \
      --restore-path .github/workflows \
      --open-gh-pr \
      --merge-message "chore: update kubebuilder scaffold" \
      --conflict-message "chore: update with conflicts - review needed"
```

## Troubleshooting

#### If you get the 403 Forbidden Error

**Error message:**
```
ERROR Update failed error=failed to open GitHub issue: gh models run failed: exit status 1
Error: unexpected response from the server: 403 Forbidden
```

**Quick fix:** Disable GitHub Models (works for everyone)

```shell
kubebuilder edit --plugins="autoupdate/v1-alpha"
```

This regenerates the workflow without GitHub Models:

```yaml
permissions:
  contents: write
  pull-requests: write
  # No issues: write or models: read unless you add --open-gh-issue / --use-gh-models

steps:
  - name: Checkout repository
    uses: actions/checkout@v4
    # ... other setup steps

  - name: Run kubebuilder alpha update
    # Default: creates and pushes the update branch and opens a Pull Request.
    # To enable AI-generated summaries, re-run:
    #   kubebuilder edit --plugins="autoupdate/v1-alpha" --use-gh-models
    run: |
      kubebuilder alpha update \
        --force \
        --push \
        --restore-path .github/workflows \
        --open-gh-pr
```

The workflow continues to work—just without AI summaries.

**To enable GitHub Models instead:**

1. Ask your GitHub administrator to enable Models (see links below)
2. Enable it in **Settings → Code and automation → Models**
3. Re-run with:

```shell
kubebuilder edit --plugins="autoupdate/v1-alpha" --use-gh-models
```

This regenerates the workflow WITH GitHub Models:

```yaml
permissions:
  contents: write
  pull-requests: write
  issues: write   # For --open-gh-issue (AI comment is posted on the issue)
  models: read    # For GitHub Models

steps:
  - name: Checkout repository
    uses: actions/checkout@v4
    # ... other setup steps

  - name: Install gh-models extension
    run: |
      gh extension install github/gh-models --force
      gh models --help >/dev/null

  - name: Run kubebuilder alpha update
    # With --use-gh-models: also creates an Issue and adds an AI-generated comment.
    run: |
      kubebuilder alpha update \
        --force \
        --push \
        --restore-path .github/workflows \
        --open-gh-pr \
        --open-gh-issue \
        --use-gh-models
```

## Demonstration

<iframe width="560" height="315" src="https://www.youtube.com/embed/dHNKx5jPSqc?si=wYwZZ0QLwFij10Sb" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

[alpha-update-command]: ./../../reference/commands/alpha_update.md
[ai-models]: https://docs.github.com/en/github-models/about-github-models
[manage-models-at-scale]: https://docs.github.com/en/github-models/github-models-at-scale/manage-models-at-scale
[manage-org-models]: https://docs.github.com/en/organizations/managing-organization-settings/managing-or-restricting-github-models-for-your-organization
