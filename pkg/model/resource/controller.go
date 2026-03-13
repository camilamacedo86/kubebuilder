/*
Copyright 2026 The Kubernetes Authors.

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

package resource

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

// Controller represents a named controller for a resource.
type Controller struct {
	// Name is the unique controller identifier within a GVK.
	Name string `json:"name,omitempty"`
}

// Validate checks that the Controller is valid.
func (c Controller) Validate() error {
	// Validate the Name
	if c.Name == "" {
		return fmt.Errorf("controller name cannot be empty")
	}

	// Controller names should be valid DNS labels
	if errors := validation.IsDNS1035Label(c.Name); len(errors) != 0 {
		return fmt.Errorf("invalid controller name %q: %s", c.Name, strings.Join(errors, ", "))
	}

	return nil
}

// Controllers holds a list of controllers for a resource.
type Controllers []Controller

// IsEmpty returns true if there are no controllers.
func (c *Controllers) IsEmpty() bool {
	return c == nil || len(*c) == 0
}

// Validate checks that all controllers are valid with unique names.
// Detects normalization collisions (e.g., "captain-backup" vs "captainbackup").
func (c *Controllers) Validate() error {
	if c.IsEmpty() {
		return nil
	}

	names := make(map[string]bool)
	normalizedNames := make(map[string]string) // normalized -> original name

	for _, controller := range *c {
		if err := controller.Validate(); err != nil {
			return err
		}

		// Check for duplicate names
		if names[controller.Name] {
			return fmt.Errorf("duplicate controller name %q", controller.Name)
		}
		names[controller.Name] = true

		// Check for normalization collisions (e.g., "captain-backup" vs "captainbackup")
		normalized := normalizeControllerName(controller.Name)
		if existingName, exists := normalizedNames[normalized]; exists {
			return fmt.Errorf("controller name %q conflicts with %q: both normalize to %q",
				controller.Name, existingName, normalized+"Reconciler")
		}
		normalizedNames[normalized] = controller.Name
	}

	return nil
}

// normalizeControllerName removes non-alphanumeric chars and lowercases for collision detection.
func normalizeControllerName(name string) string {
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		}
	}
	return strings.ToLower(result.String())
}

// NormalizeFileName converts a controller name to a valid Go filename.
// Replaces hyphens with underscores for Go file naming conventions.
func NormalizeFileName(controllerName string) string {
	return strings.ReplaceAll(controllerName, "-", "_")
}

// NormalizeReconcilerName converts a controller name to PascalCase for the reconciler struct name.
// For backwards compatibility, returns "{Kind}Reconciler" if name matches lowercase kind.
func NormalizeReconcilerName(controllerName, kind string) string {
	// Backwards compatible: no controller name or controller name matches kind
	if controllerName == "" || controllerName == strings.ToLower(kind) {
		return kind + "Reconciler"
	}

	// Convert controller name (e.g., "captain-backup") to PascalCase (e.g., "CaptainBackup")
	parts := strings.Split(controllerName, "-")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]) + part[1:])
		}
	}
	return result.String() + "Reconciler"
}

// GetControllerName returns the controller runtime name used in Named() and error logs.
// For multigroup projects, the group name is prefixed.
func GetControllerName(controllerName, kind, group string, multiGroup bool) string {
	var name string
	if controllerName != "" {
		name = controllerName
	} else {
		name = strings.ToLower(kind)
	}

	// Multigroup: prefix with group name
	if multiGroup && group != "" {
		return strings.ToLower(group) + "-" + name
	}

	return name
}

// HasController returns true if a controller with the given name exists.
func (c *Controllers) HasController(name string) bool {
	if c.IsEmpty() {
		return false
	}

	for _, controller := range *c {
		if controller.Name == name {
			return true
		}
	}
	return false
}

// AddController adds a new controller with the given name.
// Returns an error if a controller with that name already exists.
func (c *Controllers) AddController(name string) error {
	if c == nil {
		return fmt.Errorf("cannot add controller to nil Controllers")
	}

	controller := Controller{Name: name}
	if err := controller.Validate(); err != nil {
		return err
	}

	if c.HasController(name) {
		return fmt.Errorf("controller with name %q already exists", name)
	}

	*c = append(*c, controller)
	return nil
}

// GetControllerNames returns a slice of all controller names.
func (c *Controllers) GetControllerNames() []string {
	if c.IsEmpty() {
		return nil
	}

	names := make([]string, 0, len(*c))
	for _, controller := range *c {
		names = append(names, controller.Name)
	}
	return names
}

// Copy returns a deep copy of the Controllers.
func (c *Controllers) Copy() Controllers {
	if c == nil {
		return Controllers{}
	}

	controllers := make(Controllers, len(*c))
	copy(controllers, *c)
	return controllers
}

// Update combines fields of two Controllers.
// It adds controllers from other that don't exist in c.
func (c *Controllers) Update(other *Controllers) error {
	if c == nil {
		return fmt.Errorf("cannot update a nil Controllers")
	}

	if other == nil || other.IsEmpty() {
		return nil
	}

	for _, controller := range *other {
		if !c.HasController(controller.Name) {
			*c = append(*c, controller)
		}
	}

	return nil
}
