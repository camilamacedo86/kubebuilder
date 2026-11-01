//go:build integration

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

package scaffolds

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pluginutil "sigs.k8s.io/kubebuilder/v4/pkg/plugin/util"
	"sigs.k8s.io/kubebuilder/v4/test/e2e/utils"
)

var _ = Describe("create api with a file already at the controller path", func() {
	const userCode = "package controller\n\n// written by the user\n"

	var kbc *utils.TestContext

	BeforeEach(func() {
		var err error
		kbc, err = utils.NewTestContext(pluginutil.KubebuilderBinName, "GO111MODULE=on")
		Expect(err).NotTo(HaveOccurred())
		Expect(kbc.Prepare()).To(Succeed())
	})

	AfterEach(func() {
		By("removing working directory")
		kbc.Destroy()
	})

	DescribeTable("should be refused before any file is written",
		func(multiGroup bool) {
			Expect(kbc.Init("--plugins", "go/v4", "--domain", kbc.Domain)).To(Succeed())
			controllerDir := filepath.Join(kbc.Dir, "internal", "controller")
			if multiGroup {
				Expect(kbc.Edit("--multigroup=true")).To(Succeed())
				controllerDir = filepath.Join(controllerDir, kbc.Group)
			}
			controllerPath := filepath.Join(controllerDir, strings.ToLower(kbc.Kind)+"_controller.go")
			Expect(os.MkdirAll(controllerDir, 0o750)).To(Succeed())
			Expect(os.WriteFile(controllerPath, []byte(userCode), 0o600)).To(Succeed())
			before := projectFiles(kbc.Dir)

			err := kbc.CreateAPI("--group", kbc.Group, "--version", kbc.Version, "--kind", kbc.Kind,
				"--resource", "--controller", "--make=false")
			Expect(err).To(MatchError(ContainSubstring("the file already exists")))
			Expect(projectFiles(kbc.Dir)).To(Equal(before), "a refused create api must change no file")
		},
		Entry("in a single-group project", false),
		Entry("in a multigroup project", true),
	)
})

// projectFiles returns the content of every file in the project, by path. The bin directory holds
// downloaded tools, so it is skipped.
func projectFiles(dir string) map[string]string {
	GinkgoHelper()

	files := map[string]string{}
	Expect(filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == "bin" {
				return filepath.SkipDir
			}
			return nil
		}
		content, readErr := os.ReadFile(path)
		Expect(readErr).NotTo(HaveOccurred())
		files[path] = string(content)

		return nil
	})).To(Succeed())

	return files
}
