---
name: documentation-reference
description: Reference materials for Kubebuilder documentation standards
license: Apache-2.0
metadata:
  author: The Kubernetes Authors
---

# Documentation Reference

Technical references and detailed examples for Kubebuilder documentation standards.

**Documentation location**: `docs/book/src/` (built with mdBook)

This reference provides templates, examples, and patterns for contributing high-quality documentation to Kubebuilder.

## Official Style Guides

### Kubernetes Documentation Style Guide

Primary reference for all Kubebuilder documentation.

URL: https://kubernetes.io/docs/contribute/style/style-guide/

Key topics:
- Language and grammar
- Formatting standards
- Content organization
- Code examples
- Terminology

## Related Projects

### controller-runtime

Core library for building Kubernetes controllers.

- **Repository**: https://github.com/kubernetes-sigs/controller-runtime
- **Documentation**: https://pkg.go.dev/sigs.k8s.io/controller-runtime
- **When documenting**: Verify features by checking source code (see Verifying Dependency Information below)

### controller-tools

Tools for generating CRDs, webhooks, and RBAC from Go code.

- **Repository**: https://github.com/kubernetes-sigs/controller-tools
- **Documentation**: https://pkg.go.dev/sigs.k8s.io/controller-tools
- **Markers reference**: https://book.kubebuilder.io/reference/markers.html
- **When documenting**: Verify marker behavior by checking source code (see Verifying Dependency Information below)

## Documentation Build System

### mdBook

Kubebuilder documentation uses [mdBook](https://rust-lang.github.io/mdBook/) with custom preprocessors:

- **literatego** (`./litgo.sh`): Processes `{{#literatego path/to/file.go}}` includes
- **markerdocs** (`./markerdocs.sh`): Generates marker documentation

Configuration: `docs/book/book.toml`

### Building Locally

```bash
# Install mdBook (one time)
cd docs/book && ./install-and-build.sh

# Build docs
cd docs/book && mdbook build

# Serve with live reload
cd docs/book && mdbook serve
```

## Sample Generation

### Two Sources of Generated Files

Testdata projects are generated from **two different sources**:

#### 1. Tutorial-Specific Generators

**Location**: `hack/docs/internal/`

**Key files**:
- `generate.sh`: Main generation script called by `make generate-docs`
- `internal/cronjob-tutorial/`: CronJob tutorial generator
- `internal/getting-started/`: Getting Started generator
- `internal/multiversion-tutorial/`: Multiversion generator

**Generates**: Tutorial-specific customizations and examples

#### 2. Plugin Default Scaffold

**Location**: `pkg/plugins/`

**Key templates**:
- `pkg/plugins/golang/v4/scaffolds/internal/templates/agents.go` - Generates `AGENTS.md`
- `pkg/plugins/golang/v4/scaffolds/internal/templates/` - Other default boilerplate
- `pkg/plugins/optional/*/scaffolds/internal/templates/` - Optional plugin templates

**Generates**: Default project structure files (Makefile, main.go, AGENTS.md, etc.)

### Critical Workflow for Fixing Testdata

**If you find a documentation issue in testdata files**:

1. **Identify the source**:
   - Tutorial-specific content (custom examples, tutorial text) → Edit `hack/docs/internal/`
   - Default scaffold files (`AGENTS.md`, standard boilerplate) → Edit `pkg/plugins/`

2. **Fix the template source**:
   - Never edit generated files directly
   - Always edit the template that generates them

3. **Rebuild and regenerate**:
   ```bash
   make install         # Rebuild kubebuilder binary with template changes
   make generate-docs   # Regenerate all testdata using new binary
   ```

4. **Verify the fix**:
   ```bash
   git diff docs/book/src/*/testdata/project/  # Check generated files updated correctly
   ```

### Generated Locations

Auto-generated (DO NOT EDIT DIRECTLY):
- `docs/book/src/cronjob-tutorial/testdata/project/`
- `docs/book/src/getting-started/testdata/project/`
- `docs/book/src/multiversion-tutorial/testdata/project/`

### Commands

```bash
# Regenerate all samples (after template changes)
make install         # Required after changing pkg/plugins/ templates
make generate-docs   # Regenerates all testdata

# Fix accessibility and trailing spaces
make fix-docs

# Test tutorial code compiles
make test-book
```

## Content Structure Templates

### Plugin Documentation Structure

Each plugin in `docs/book/src/plugins/available/` should follow this structure:

```markdown
# Plugin Name (plugin-key/version)

Brief description of what the plugin does and the value it provides.

By using this plugin, you will get:
- Feature 1
- Feature 2
- Feature 3

<aside class="note" role="note">
<p class="note-title">Examples</p>

See the `project-v4-with-plugins` directory under the [testdata][testdata]
directory in the Kubebuilder project to check an example
of scaffolding created using this plugin.

Example scaffolding command:
```shell
kubebuilder create api \
  --group example.com \
  --version v1alpha1 \
  --kind MyKind \
  --plugins="plugin-name/version"
```
</aside>

## When to use it?

- Use case 1
- Use case 2
- Use case 3

## How to use it?

1. **Initialize your project**:
   After creating a new project with `kubebuilder init`, you can use this plugin...

2. **Create APIs** (or relevant step):
   Example command with flags explained:
   ```sh
   kubebuilder create api --group example.com --version v1alpha1 --kind MyKind --plugins="plugin-name/version"
   ```

3. **Additional steps as needed**

## Examples

Include from testdata:
{{#include ../../testdata/project-v4-with-plugins/path/to/file.go}}

## Configuration

Any plugin-specific configuration details.

## Limitations

Known limitations or constraints.

## Further resources

Links to related documentation.

[testdata]: https://github.com/kubernetes-sigs/kubebuilder/tree/master/testdata
[controller-runtime]: https://github.com/kubernetes-sigs/controller-runtime
```

### Tutorial Structure

```markdown
# Tutorial Title

Brief introduction explaining what the reader will learn.

## Prerequisites

- Required knowledge
- Required tools and versions
- Required setup

## Step 1: Action verb describing the step

Explanation of what this step accomplishes.

[Commands or code]

Expected output or result.

## Step 2: Next action

...

## What's next

- Related tutorials
- Advanced topics
- Reference documentation
```

### Conceptual Documentation

```markdown
# Concept Title

One-paragraph overview.

## What is [concept]

Clear definition and explanation.

## Why [concept] matters

Use cases and benefits.

## How [concept] works

Technical details and architecture.

## Related concepts

Links to related documentation.
```

### Reference Documentation

```markdown
# API/Tool Reference

Brief description.

## Synopsis

Command syntax or API signature.

## Description

Detailed explanation.

## Options/Parameters

Table or list of all options.

## Examples

Common usage examples, simple to complex.

## See also

Related commands or APIs.
```

## Language Examples

### Plain English

Good:
```markdown
The controller watches for changes to resources.
```

Avoid:
```markdown
The controller leverages an event-driven architecture to facilitate real-time synchronization of resource state.
```

### Active Voice

Good:
```markdown
Kubebuilder generates the boilerplate code.
```

Avoid:
```markdown
The boilerplate code is generated by Kubebuilder.
```

### Present Tense

Good:
```markdown
The webhook validates the resource.
```

Avoid:
```markdown
The webhook will validate the resource.
```

### Direct Address

Good:
```markdown
Run the following command to create a new API.
```

Avoid:
```markdown
We can run the following command to create a new API.
One should run the following command to create a new API.
```

### No Contractions

Good:
```markdown
The controller does not support external resources.
You will see the results in the cluster.
```

Avoid:
```markdown
The controller doesn't support external resources.
You'll see the results in the cluster.
```

### Show, Don't Tell

Good (shows with example):
```markdown
The controller reconciles the resource state:

```go
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Fetch the custom resource
    obj := &myv1.MyKind{}
    if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Update status to reflect current state
    obj.Status.ObservedGeneration = obj.Generation
    return ctrl.Result{}, r.Status().Update(ctx, obj)
}
```

The reconciler fetches the resource, updates its status, and returns.
```

Avoid (tells without showing):
```markdown
The controller has a reconciliation function that is responsible for managing the lifecycle of resources. It performs operations to ensure the desired state matches the actual state by fetching resources from the API server and updating their status accordingly.
```

**Why**: Code examples demonstrate actual behavior better than abstract descriptions.

## Formatting Examples

### Sentence Case Headings

Good:
```markdown
## Creating a new controller
```

Avoid:
```markdown
## Creating A New Controller
```

### Code Blocks with Language

Good:
````markdown
```go
func main() {
    // your code
}
```
````

Avoid:
````markdown
```
func main() {
    // your code
}
```
````

### Inline Code Usage

Use inline code for:
- Commands: `make install`
- File names: `main.go`
- API objects: `Deployment`
- Field names: `spec.replicas`
- Short snippets: `return err`

Use code blocks for:
- Multi-line code
- Command output
- File contents
- Configuration files

## Link Examples

### Internal Links with Relative Paths

Good:
```markdown
See [Creating a controller](../cronjob-tutorial/controller-implementation.md) for details.
See [API markers](./markers.md) in the same directory.
```

Avoid:
```markdown
See [guide](/docs/book/src/getting-started.md) for details.
See [guide](https://kubebuilder.io/getting-started.md) for internal content.
```

### Link Aliases at Bottom

```markdown
Content with links to [Kubernetes API][k8s-api] and [contributing guide][contributing].

[k8s-api]: https://kubernetes.io/docs/reference/
[contributing]: ../CONTRIBUTING.md
```

### Descriptive Link Text

Good:
```markdown
See the [Kubernetes API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md).
```

Avoid:
```markdown
Click [here](https://github.com/...) for API conventions.
```

## Code Example Patterns

### CRITICAL: Always Use Testdata Includes When Available

**DO NOT copy/paste code inline if it exists in testdata.** Always use includes when the code is available in testdata projects.

### Why This Matters

**Problem with inline code**:
```markdown
# BAD: This will become outdated
```go
func (r *CronJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)
    // ... 50 lines of code copied from somewhere
}
```
```

When the code changes, the documentation becomes **wrong**. Users copy broken examples.

**Solution with testdata includes**:
```markdown
# GOOD: This stays current automatically
The reconciliation logic handles resource lifecycle:

{{#literatego ./testdata/project/internal/controller/cronjob_controller.go}}

The reconciler fetches the CronJob, manages child Jobs, and updates status.
```

When code changes, **documentation auto-updates**. Users always get working examples.

### Available Testdata

**Tutorial projects** (auto-generated, tested):
- `docs/book/src/cronjob-tutorial/testdata/project/`
- `docs/book/src/getting-started/testdata/project/`
- `docs/book/src/multiversion-tutorial/testdata/project/`

**Plugin examples** (tested):
- `testdata/project-v4-with-plugins/`

All tested by `make test-book` - guaranteed to compile and work.

### Include Shortcode Syntax

**For Go code** (adds syntax highlighting, handles imports):
```markdown
{{#literatego ./testdata/project/path/to/file.go}}
```

**For YAML, JSON, Makefile, shell** (includes raw):
```markdown
{{#include ./testdata/project/config/manager/manager.yaml}}
```

**For specific sections** (using anchors in source files):
```markdown
{{#include ./testdata/project/config/default/kustomization.yaml:webhook-resources}}
```

### Using Anchors in Source Files

Add anchors to mark sections:

```go
// +kubebuilder:docs-gen:collapse=Apache License

// ANCHOR: imports
import (
    "context"
    ctrl "sigs.k8s.io/controller-runtime"
)
// ANCHOR_END: imports
```

Then include just that section:
```markdown
{{#literatego ./testdata/project/main.go:imports}}
```

### Documentation Around Includes

**Pattern** (always follow this):
1. **Context before**: Explain what the reader will see
2. **Include shortcode**: Reference testdata
3. **Explanation after**: Summarize what they saw

**Good example**:
```markdown
## Implementing the controller

The basic logic of our CronJob controller is this:

1. Load the named CronJob
2. List all active jobs, and update the status
3. Clean up old jobs according to the history limits

{{#literatego ./testdata/project/internal/controller/cronjob_controller.go}}

The reconciler returns successfully when the CronJob is deleted (using `client.IgnoreNotFound`).
For existing CronJobs, it ensures child Jobs match the schedule.
```

**Why this pattern works**:
- Reader knows what to look for (context)
- Sees actual, tested code (include)
- Understands the key points (explanation)

### When You Cannot Use Includes

If code does not exist in testdata:

1. **First**: Consider adding it to testdata (best option)
2. **Second**: Use very short inline snippet (1-5 lines max)
3. **Third**: Create GitHub issue to add to testdata
4. **Never**: Copy/paste large blocks of code inline

**Example of acceptable inline** (pattern only, not real code):
```markdown
## Custom validation

Add validation to your webhook:

```go
if obj.Spec.Replicas < 0 {
    return fmt.Errorf("replicas must be non-negative")
}
```

This ensures the replica count is valid before saving.
```

### Inline Code Comments

```go
// Good: Comments explain WHY, not WHAT
// Reconcile implements the control loop logic.
// It ensures the actual state matches the desired state.
func (r *MyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Fetch the instance
    obj := &myv1.MyKind{}
    if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Your logic here
    return ctrl.Result{}, nil
}
```

### Shell Examples

Use `$` for commands:
```bash
$ kubebuilder init --domain example.com
$ make install
```

No `$` for output:
```bash
$ kubectl get pods
NAME                     READY   STATUS    RESTARTS   AGE
my-pod-abc123           1/1     Running   0          5m
```

## Admonitions (Aside Blocks)

### Note Blocks

Use for important information:
```markdown
<aside class="note" role="note">
<p class="note-title">Title Here</p>

Content goes here. Can include markdown, code blocks, links, etc.

</aside>
```

### Warning Blocks

Use for critical information or caveats:
```markdown
<aside class="warning" role="note">
<p class="note-title">Title Here</p>

Content goes here. Can include markdown, code blocks, links, etc.

</aside>
```

### General Guidelines

- Always include `role="note"` attribute
- Use descriptive titles with `<p class="note-title">`
- Leave a blank line after the title
- Can contain multiple paragraphs, code blocks, lists
- Always close with blank line before `</aside>`
- Do NOT use "TIP!" or similar callouts in regular text - use aside blocks only

## Capitalization Standards

### Technical Terms in Prose

**Always capitalize:**
```markdown
Kubernetes provides a declarative API.
The controller watches YAML files for changes.
Export the configuration as JSON.
Use the HTTP endpoint for health checks.
```

**Always lowercase:**
```markdown
Create a new namespace for the controller.
The kubectl command interacts with the cluster.
Install kubebuilder on your system.
Use kustomize to manage configurations.
```

### Why This Matters

Consistent capitalization:
- Improves readability for non-native English speakers
- Distinguishes proper nouns (Kubernetes) from common nouns (cluster)
- Follows Kubernetes community standards
- Aligns with official product names (YAML, not yaml)

## Accessibility Guidelines

Requirements:
- Semantic HTML headings (h1, h2, h3)
- Alt text for images
- Descriptive link text
- Proper heading hierarchy
- Sufficient color contrast
- No trailing spaces
- Proper aside block format (not deprecated shortcodes)

Run `make fix-docs` to automatically fix accessibility issues and remove trailing spaces.

## Common External References

Centralize frequently used URLs:

- Kubernetes documentation: https://kubernetes.io/docs/
- Kubernetes API conventions: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
- Kubernetes glossary: https://kubernetes.io/docs/reference/glossary/
- mdBook documentation: https://rust-lang.github.io/mdBook/
- Go documentation: https://go.dev/doc/

**Kubebuilder ecosystem**:
- controller-runtime docs: https://pkg.go.dev/sigs.k8s.io/controller-runtime
- controller-runtime repo: https://github.com/kubernetes-sigs/controller-runtime
- controller-tools docs: https://pkg.go.dev/sigs.k8s.io/controller-tools
- controller-tools repo: https://github.com/kubernetes-sigs/controller-tools

## Target Audience Levels

- **Getting Started**: Assumes basic Kubernetes knowledge
- **Tutorials**: Step-by-step with explanations
- **Reference**: Complete technical details
- **Conceptual**: Architectural overview

## Verifying Dependency Information

When documenting features from controller-runtime, controller-tools, or other dependencies, verify accuracy by checking the source code.

### Why This Matters

**Problem**: Documentation about dependencies can become outdated or incorrect:
- API methods change signatures
- Marker behavior changes
- Default values change
- New options are added

**Solution**: Verify by checking the actual source code.

### How to Verify

Use `go mod vendor` to download and inspect dependency source code:

```bash
# Create temp directory for verification
mkdir -p /tmp/verify-deps
cd /tmp/verify-deps

# Initialize a Go module
go mod init example.com/verify

# Add the dependency you want to check
go get sigs.k8s.io/controller-runtime@latest
# or
go get sigs.k8s.io/controller-tools@latest

# Vendor the dependencies
go mod vendor

# Now inspect the source code
cd vendor/sigs.k8s.io/controller-runtime
# or
cd vendor/sigs.k8s.io/controller-tools
```

### What to Verify

**For controller-runtime**:
- Manager options and defaults: `vendor/sigs.k8s.io/controller-runtime/pkg/manager/manager.go`
- Client behavior: `vendor/sigs.k8s.io/controller-runtime/pkg/client/client.go`
- Reconciler interface: `vendor/sigs.k8s.io/controller-runtime/pkg/reconcile/reconcile.go`
- Builder options: `vendor/sigs.k8s.io/controller-runtime/pkg/builder/controller.go`

**For controller-tools**:
- Marker definitions: `vendor/sigs.k8s.io/controller-tools/pkg/markers/`
- CRD generation: `vendor/sigs.k8s.io/controller-tools/pkg/crd/`
- Webhook generation: `vendor/sigs.k8s.io/controller-tools/pkg/webhook/`
- RBAC generation: `vendor/sigs.k8s.io/controller-tools/pkg/rbac/`

### Example Verification Workflow

Documenting a controller-runtime feature:

```bash
# Setup
mkdir -p /tmp/verify-controller-runtime
cd /tmp/verify-controller-runtime
go mod init example.com/verify
go get sigs.k8s.io/controller-runtime@latest
go mod vendor

# Check Manager.Start behavior
cat vendor/sigs.k8s.io/controller-runtime/pkg/manager/manager.go | grep -A 10 "func.*Start"

# Check default reconcile options
grep -r "DefaultRecoverPanic" vendor/sigs.k8s.io/controller-runtime/

# Verify the exact version being documented
go list -m sigs.k8s.io/controller-runtime
```

### When to Verify

Always verify when:
- Documenting specific API behavior
- Describing default values or configurations
- Explaining marker syntax or options
- Updating documentation for new versions
- User reports documentation is incorrect

Do NOT document from memory or assumptions - verify the source code.

## Testing Documentation

Before submitting:

1. **Test all command examples manually** in a clean environment:
   ```bash
   # Create temp directory for testing
   mkdir -p /tmp/kb-doc-test
   cd /tmp/kb-doc-test

   # Run commands exactly as documented
   kubebuilder init --domain example.com
   # ... etc

   # Verify output matches documentation
   ```

2. **Verify code examples compile and run**:
   - Go code: Ensure it compiles without errors
   - YAML/JSON: Validate syntax
   - Shell scripts: Test in clean shell

3. **Check all links are valid**:
   - Internal links resolve correctly
   - External links are accessible
   - No broken anchors

4. **Test tutorial code**: `make test-book` (compiles all testdata examples)

5. **Regenerate samples if changed**: `make generate-docs`

6. **Fix formatting**: `make fix-docs` (trailing spaces, accessibility)

7. **Build docs locally**: `cd docs/book && mdbook serve`

8. **Review rendered output** for formatting issues

### Why Testing Matters

Untested examples lead to:
- Users unable to follow tutorials
- Loss of trust in documentation
- Support burden from broken examples
- Negative project perception

**Rules**:
1. If you cannot test an example yourself, do not include it
2. If example code exists in testdata, ALWAYS use `{{#include}}` - never copy/paste
3. If code does not exist in testdata, you can use short inline examples (1-5 lines)
4. For larger examples not in testdata, consider adding them to testdata first

### Maintainability Impact

| Approach | Maintainability | Accuracy | User Trust |
|----------|----------------|----------|------------|
| Testdata includes (`{{#include}}`) | Auto-updates | Tested | High |
| Inline code examples | Manual updates needed | Drifts over time | Low |

**Real scenario**:
- 2024: CronJob tutorial uses inline code
- 2025: controller-runtime API changes
- Result: Tutorial has broken, outdated code
- Users: Cannot follow tutorial, file bugs

**With includes**:
- 2024: CronJob tutorial uses `{{#literatego}}`
- 2025: Testdata updated, `make test-book` passes
- Result: Documentation automatically current
- Users: Tutorial works perfectly

## Review Patterns for AI

### How to Provide Feedback

When reviewing documentation, provide structured feedback:

**Format:**
```
**Issue**: Brief description of the problem
**Current**: Show the problematic text
**Suggested**: Show the corrected text
**Reason**: Explain why (reference standard)
```

**Example:**
```
**Issue**: Uses contraction in technical explanation
**Current**: "The controller doesn't validate resources"
**Suggested**: "The controller does not validate resources"
**Reason**: No contractions in technical documentation (Language Standards)
```

### Batch Review Approach

For multiple similar issues:
1. Identify the pattern (e.g., all contractions, all passive voice)
2. Provide 2-3 examples with full feedback
3. Summarize remaining instances: "Found 12 more instances of contractions in lines X, Y, Z..."
4. Suggest running automated tools when applicable

### Priority Order

Focus on issues in this order:
1. **Correctness**: Wrong commands, broken links, incorrect code
2. **Clarity**: Confusing explanations, missing context
3. **Style**: Voice, tense, contractions, capitalization
4. **Formatting**: Code block tags, heading case, link format
