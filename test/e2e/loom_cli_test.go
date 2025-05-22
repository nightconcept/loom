// Package e2e contains end-to-end tests for the Loom CLI tool.
package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Loom CLI", func() {
	var loomExecutable string

	BeforeEach(func() {
		basePath, err := filepath.Abs("../..")
		Expect(err).NotTo(HaveOccurred())

		if runtime.GOOS == "windows" {
			loomExecutable = filepath.Join(basePath, "build", "loom.exe")
		} else {
			loomExecutable = filepath.Join(basePath, "build", "loom")
		}

		Expect(loomExecutable).To(BeAnExistingFile(), "Loom executable not found at "+loomExecutable+". Make sure to build it before running tests.")
	})

	Describe("Basic CLI functionality", func() {
		Context("when running 'loom' with no arguments", func() {
			It("should output help information", func() {
				command := exec.Command(loomExecutable)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))

				output := session.Buffer()
				Expect(output).To(gbytes.Say("loom"))
				Expect(output).To(Or(
					gbytes.Say("USAGE:"),
					gbytes.Say("Usage:"),
					gbytes.Say("Commands:"),
					gbytes.Say("options:"),
				))
			})
		})
	})

	Describe("loom init functionality", func() {
		var tempTestDir string

		BeforeEach(func() {
			tempTestDir = CreateTempDir()
		})

		Context("when 'loom.yaml' does not exist", func() {
			It("should create 'loom.yaml' with default content and print a success message", func() {
				command := exec.Command(loomExecutable, "init")
				command.Dir = tempTestDir

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))

				Expect(session.Out).To(gbytes.Say("Initialized empty Loom project with loom.yaml"))

				loomYAMLPath := filepath.Join(tempTestDir, "loom.yaml")
				Expect(loomYAMLPath).To(BeAnExistingFile())

				yamlContent, err := os.ReadFile(loomYAMLPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(yamlContent)).To(ContainSubstring("# loom.yaml - Loom project configuration file"))
				Expect(string(yamlContent)).To(ContainSubstring("version: \"1\""))
				Expect(string(yamlContent)).To(ContainSubstring("threads: []"))
			})
		})

		Context("when 'loom.yaml' exists and is empty", func() {
			BeforeEach(func() {
				err := os.WriteFile(filepath.Join(tempTestDir, "loom.yaml"), []byte{}, 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should overwrite 'loom.yaml' with default content and print a success message", func() {
				command := exec.Command(loomExecutable, "init")
				command.Dir = tempTestDir

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))

				Expect(session.Out).To(gbytes.Say("Initialized empty Loom project with loom.yaml"))

				loomYAMLPath := filepath.Join(tempTestDir, "loom.yaml")
				yamlContent, err := os.ReadFile(loomYAMLPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(yamlContent)).To(ContainSubstring("version: \"1\""))
				Expect(string(yamlContent)).To(ContainSubstring("threads: []"))
			})
		})

		Context("when 'loom.yaml' exists and contains only comments and whitespace", func() {
			BeforeEach(func() {
				content := "# This is a comment\n   \n# Another comment\n"
				err := os.WriteFile(filepath.Join(tempTestDir, "loom.yaml"), []byte(content), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should overwrite 'loom.yaml' with default content and print a success message", func() {
				command := exec.Command(loomExecutable, "init")
				command.Dir = tempTestDir

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))

				Expect(session.Out).To(gbytes.Say("Initialized empty Loom project with loom.yaml"))

				loomYAMLPath := filepath.Join(tempTestDir, "loom.yaml")
				yamlContent, err := os.ReadFile(loomYAMLPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(yamlContent)).To(ContainSubstring("version: \"1\""))
				Expect(string(yamlContent)).To(ContainSubstring("threads: []"))
			})
		})

		Context("when 'loom.yaml' exists and is not empty", func() {
			var existingContent string
			BeforeEach(func() {
				existingContent = "version: \"1\"\nthreads:\n  - name: existingThread\n    source: someSource"
				err := os.WriteFile(filepath.Join(tempTestDir, "loom.yaml"), []byte(existingContent), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should fail, print an error message, and not modify the existing file", func() {
				command := exec.Command(loomExecutable, "init")
				command.Dir = tempTestDir

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1))

				Expect(session.Err).To(gbytes.Say("failed to initialize project: loom.yaml already exists and is not empty"))

				loomYAMLPath := filepath.Join(tempTestDir, "loom.yaml")
				yamlContent, err := os.ReadFile(loomYAMLPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(yamlContent)).To(Equal(existingContent))
			})
		})

		Context("when 'loom init' is run with extraneous arguments", func() {
			It("should ignore the arguments and initialize the project successfully", func() {
				command := exec.Command(loomExecutable, "init", "extraneousArg1", "--someflag")
				command.Dir = tempTestDir

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))

				Expect(session.Out).To(gbytes.Say("Initialized empty Loom project with loom.yaml"))

				loomYAMLPath := filepath.Join(tempTestDir, "loom.yaml")
				Expect(loomYAMLPath).To(BeAnExistingFile())
				yamlContent, err := os.ReadFile(loomYAMLPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(yamlContent)).To(ContainSubstring("version: \"1\""))
			})
		})
	})

	Describe("loom add functionality", func() {
		var tempProjectDir string
		var tempGlobalLoomDir string
		var originalLoomGlobalDirEnv string
		var mockStorePath string

		BeforeEach(func() {
			tempProjectDir = CreateTempDir()
			tempGlobalLoomDir = CreateTempDir()
			originalLoomGlobalDirEnv, _ = os.LookupEnv("LOOM_GLOBAL_DIR")
			mockStorePath = filepath.Join(tempGlobalLoomDir, "myStore")
			err := os.MkdirAll(mockStorePath, 0755)
			Expect(err).NotTo(HaveOccurred())

			globalConfigContent := `
version: "1"
stores:
  - name: myStore
    type: local
    path: "` + filepath.ToSlash(mockStorePath) + `"
`
			err = os.WriteFile(filepath.Join(tempGlobalLoomDir, "loom.yaml"), []byte(globalConfigContent), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			var err error
			if originalLoomGlobalDirEnv == "" {
				err = os.Unsetenv("LOOM_GLOBAL_DIR")
			} else {
				err = os.Setenv("LOOM_GLOBAL_DIR", originalLoomGlobalDirEnv)
			}
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when adding a thread from a configured local store (happy path)", func() {
			It("should copy thread files to the project and update project loom.yaml", func() {
				mockThreadName := "myTestThread"
				mockThreadSourceDir := filepath.Join(mockStorePath, mockThreadName, "_thread")
				err := os.MkdirAll(mockThreadSourceDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				CreateTempFile(mockThreadSourceDir, "file1.txt", "content of file1")
				CreateTempFile(filepath.Join(mockThreadSourceDir, "subdir"), "file2.txt", "content of file2")

				command := exec.Command(loomExecutable, "add", mockThreadName)
				command.Dir = tempProjectDir

				env := []string{}
				for _, e := range os.Environ() {
					if !strings.HasPrefix(e, "LOOM_GLOBAL_DIR=") {
						env = append(env, e)
					}
				}
				command.Env = append(env, "LOOM_GLOBAL_DIR="+tempGlobalLoomDir)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))

				Expect(session.Out).To(gbytes.Say("Thread 'myTestThread' added successfully from myStore"))

				Expect(filepath.Join(tempProjectDir, "file1.txt")).To(BeAnExistingFile())
				Expect(filepath.Join(tempProjectDir, "subdir", "file2.txt")).To(BeAnExistingFile())

				projectLoomYAMLPath := filepath.Join(tempProjectDir, "loom.yaml")
				Expect(projectLoomYAMLPath).To(BeAnExistingFile())
				yamlContent, err := os.ReadFile(projectLoomYAMLPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(yamlContent)).To(ContainSubstring("name: " + mockThreadName))
				Expect(string(yamlContent)).To(ContainSubstring("source: myStore"))
				Expect(string(yamlContent)).To(ContainSubstring("./:"))
				Expect(string(yamlContent)).To(ContainSubstring("- file1.txt"))
				Expect(string(yamlContent)).To(ContainSubstring("subdir/:"))
				Expect(string(yamlContent)).To(ContainSubstring("- file2.txt"))
			})
		})

		Context("when adding a thread that is malformed (e.g., _thread is a file)", func() {
			It("should output an error and not add the thread", func() {
				mockThreadName := "malformedThread"
				mockThreadDir := filepath.Join(mockStorePath, mockThreadName)
				err := os.MkdirAll(mockThreadDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				CreateTempFile(mockThreadDir, "_thread", "this is a file, not a directory")

				command := exec.Command(loomExecutable, "add", mockThreadName)
				command.Dir = tempProjectDir

				env := []string{}
				for _, e := range os.Environ() {
					if !strings.HasPrefix(e, "LOOM_GLOBAL_DIR=") {
						env = append(env, e)
					}
				}
				command.Env = append(env, "LOOM_GLOBAL_DIR="+tempGlobalLoomDir)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(1))

				rawExpectedErrorMsg := "thread path '" + filepath.Join(mockStorePath, mockThreadName, "_thread") + "' in store 'myStore' is a file, not a directory"
				expectedErrorMsgForMatcher := regexp.QuoteMeta(rawExpectedErrorMsg)
				Expect(session.Err).To(gbytes.Say(expectedErrorMsgForMatcher))

				Expect(filepath.Join(tempProjectDir, "file1.txt")).NotTo(BeAnExistingFile())
				Expect(filepath.Join(tempProjectDir, "_thread")).NotTo(BeAnExistingFile())

				projectLoomYAMLPath := filepath.Join(tempProjectDir, "loom.yaml")
				if _, err := os.Stat(projectLoomYAMLPath); err == nil {
					yamlContent, readErr := os.ReadFile(projectLoomYAMLPath)
					Expect(readErr).NotTo(HaveOccurred())
					Expect(string(yamlContent)).NotTo(ContainSubstring("name: " + mockThreadName))
				} else {
					Expect(os.IsNotExist(err)).To(BeTrue(), "loom.yaml should not exist or error should be IsNotExist")
				}
			})
		})
	})

	Describe("loom add command E2E Test Scenarios", func() {
		var tempProjectDir string
		var tempGlobalLoomDir string
		var originalLoomGlobalDirEnv string
		var loomExecPath string

		BeforeEach(func() {
			basePath, err := filepath.Abs("../..")
			Expect(err).NotTo(HaveOccurred())

			if runtime.GOOS == "windows" {
				loomExecPath = filepath.Join(basePath, "build", "loom.exe")
			} else {
				loomExecPath = filepath.Join(basePath, "build", "loom")
			}
			Expect(loomExecPath).To(BeAnExistingFile(), "Loom executable not found at "+loomExecPath)

			tempProjectDir = CreateTempDir()
			tempGlobalLoomDir = CreateTempDir()
			originalLoomGlobalDirEnv, _ = os.LookupEnv("LOOM_GLOBAL_DIR")

			InitProjectLoomFile(tempProjectDir)
		})

		AfterEach(func() {
			var err error
			if originalLoomGlobalDirEnv == "" {
				err = os.Unsetenv("LOOM_GLOBAL_DIR")
			} else {
				err = os.Setenv("LOOM_GLOBAL_DIR", originalLoomGlobalDirEnv)
			}
			Expect(err).NotTo(HaveOccurred())
		})

		runLoomAdd := func(args ...string) *gexec.Session {
			command := exec.Command(loomExecPath, append([]string{"add"}, args...)...)
			command.Dir = tempProjectDir
			env := os.Environ()
			filteredEnv := []string{}
			for _, e := range env {
				if !strings.HasPrefix(e, "LOOM_GLOBAL_DIR=") {
					filteredEnv = append(filteredEnv, e)
				}
			}
			command.Env = append(filteredEnv, "LOOM_GLOBAL_DIR="+tempGlobalLoomDir)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			return session
		}

		Describe("Argument Parsing", func() {
			Context("when running 'loom add' with no arguments", func() {
				It("should fail with a usage message", func() {
					session := runLoomAdd()
					Eventually(session).Should(gexec.Exit(1))
					Expect(session.Err).To(gbytes.Say("thread name or store/thread is required"))
					Expect(session.Out.Contents()).To(BeEmpty())
				})
			})

			Context("when running 'loom add /'", func() {
				It("should fail due to invalid format (empty store and thread name)", func() {
					session := runLoomAdd("/")
					Eventually(session).Should(gexec.Exit(1))
					Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("invalid format for store/thread: '/'. Both store name and thread name must be specified")))
				})
			})

			Context("when running 'loom add store/'", func() {
				It("should fail due to invalid format (missing thread name)", func() {
					session := runLoomAdd("storeName/")
					Eventually(session).Should(gexec.Exit(1))
					Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("invalid format for store/thread: 'storeName/'. Both store name and thread name must be specified")))
				})
			})

			Context("when running 'loom add /thread'", func() {
				It("should fail due to invalid format (missing store name)", func() {
					session := runLoomAdd("/threadName")
					Eventually(session).Should(gexec.Exit(1))
					Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("invalid format for store/thread: '/threadName'. Both store name and thread name must be specified")))
				})
			})
		})

		Describe("Thread Source and Resolution", func() {
			var (
				projectLoomDir string
			)

			BeforeEach(func() {
				projectLoomDir = filepath.Join(tempProjectDir, ".loom")
				err := os.MkdirAll(projectLoomDir, 0755)
				Expect(err).NotTo(HaveOccurred())
				InitProjectLoomFile(tempProjectDir)
			})

			Context("when a thread with the same name already exists in the project's .loom directory", func() {
				It("should inform the user and not overwrite the existing thread", func() {
					threadName := "myExistingThread"
					threadPath := filepath.Join(projectLoomDir, threadName)
					threadSourcePath := filepath.Join(threadPath, "_thread")
					err := os.MkdirAll(threadSourcePath, 0755)
					Expect(err).NotTo(HaveOccurred())

					sampleFilePath := filepath.Join(threadSourcePath, "sample.txt")
					err = os.WriteFile(sampleFilePath, []byte("This is a local thread."), 0644)
					Expect(err).NotTo(HaveOccurred())

					session := runLoomAdd(threadName)
					Eventually(session).Should(gexec.Exit(0))
					Expect(session.Out).To(gbytes.Say(regexp.QuoteMeta("Thread '" + threadName + "' added successfully from project:.loom/" + threadName)))

					projectFilePath := filepath.Join(tempProjectDir, "sample.txt")
					Expect(projectFilePath).To(BeAnExistingFile())
					content, err := os.ReadFile(projectFilePath)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(content)).To(Equal("This is a local thread."))
				})
			})

			Context("when adding a thread by specifying an existing store name (e.g., loom add myStore/myTestThread)", func() {
				var storePath string
				BeforeEach(func() {
					storePath = filepath.Join(tempGlobalLoomDir, "myStore")
					err := os.MkdirAll(storePath, 0755)
					Expect(err).NotTo(HaveOccurred())

					threadName := "myTestThread"
					threadPath := filepath.Join(storePath, threadName)
					threadSourcePath := filepath.Join(threadPath, "_thread")

					err = os.MkdirAll(threadSourcePath, 0755)
					Expect(err).NotTo(HaveOccurred())

					sampleFilePath := filepath.Join(threadSourcePath, "sample.txt")
					err = os.WriteFile(sampleFilePath, []byte("Content from myStore."), 0644)
					Expect(err).NotTo(HaveOccurred())

					globalLoomConfigPath := filepath.Join(tempGlobalLoomDir, "loom.yaml")
					configContent := fmt.Sprintf("version: \"1\"\nstores:\n  - name: myStore\n    type: local\n    path: %s\n", filepath.ToSlash(storePath))
					err = os.WriteFile(globalLoomConfigPath, []byte(configContent), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should successfully add the thread from the specified global store", func() {
					session := runLoomAdd("myStore/myTestThread")
					Eventually(session).Should(gexec.Exit(0))
					Expect(session.Out).To(gbytes.Say("myTestThread"))
					Expect(session.Out).To(gbytes.Say("myStore"))

					projectSamplePath := filepath.Join(tempProjectDir, "sample.txt")
					Expect(projectSamplePath).To(BeAnExistingFile())
					content, err := os.ReadFile(projectSamplePath)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(content)).To(Equal("Content from myStore."))

					projectLoomConfig, err := os.ReadFile(filepath.Join(tempProjectDir, "loom.yaml"))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(projectLoomConfig)).To(ContainSubstring("name: myTestThread"))
					Expect(string(projectLoomConfig)).To(ContainSubstring("source: myStore"))
				})
			})

			Context("when the specified thread is not found in the specified store", func() {
				var storePath string
				BeforeEach(func() {
					storePath = filepath.Join(tempGlobalLoomDir, "stores", "anotherStore")
					err := os.MkdirAll(storePath, 0755)
					Expect(err).NotTo(HaveOccurred())
					globalLoomConfigPath := filepath.Join(tempGlobalLoomDir, "loom.yaml")
					configContent := fmt.Sprintf("version: \"1\"\nstores:\n  - name: anotherStore\n    path: %s\n", storePath)
					err = os.WriteFile(globalLoomConfigPath, []byte(configContent), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should fail and report that the thread was not found in the store", func() {
					session := runLoomAdd("anotherStore/nonExistentThread")
					Eventually(session).Should(gexec.Exit(1))
					Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("thread 'nonExistentThread' not found in specified store 'anotherStore'")))
				})
			})

			Context("when the specified store name does not exist in the global configuration", func() {
				It("should fail and report that the store is not configured", func() {
					globalLoomConfigPath := filepath.Join(tempGlobalLoomDir, "loom.yaml")
					err := os.WriteFile(globalLoomConfigPath, []byte("stores: []"), 0644)
					Expect(err).NotTo(HaveOccurred())

					session := runLoomAdd("unknownStore/anyThread")
					Eventually(session).Should(gexec.Exit(1))
					Expect(session.Err).To(gbytes.Say(regexp.QuoteMeta("store 'unknownStore' not found in global configuration")))
				})
			})

			Context("when the thread is not found in any configured store or project .loom/ folder (implicit resolution)", func() {
				BeforeEach(func() {
					globalLoomConfigPath := filepath.Join(tempGlobalLoomDir, "loom.yaml")
					err := os.WriteFile(globalLoomConfigPath, []byte("stores: []"), 0644)
					Expect(err).NotTo(HaveOccurred())
				})
				It("should fail and report that the thread could not be found", func() {
					session := runLoomAdd("completelyMissingThread")
					Eventually(session).Should(gexec.Exit(1))

					Expect(session.Err).To(gbytes.Say("completelyMissingThread"))
					Expect(session.Err).To(gbytes.Say("not found"))
				})
			})

		})

		Describe("File Conflict Handling", func() {
		})

		Describe("Project loom.yaml Manipulation", func() {
		})

		Describe("Extraneous Arguments", func() {
		})
	})
})
