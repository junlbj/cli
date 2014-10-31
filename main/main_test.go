package main_test

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("main", func() {
	var (
		old_PLUGINS_HOME string
	)

	BeforeEach(func() {
		old_PLUGINS_HOME = os.Getenv("CF_PLUGIN_HOME")

		dir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		fullDir := filepath.Join(dir, "..", "fixtures", "config", "main-plugin-test-config")
		err = os.Setenv("CF_PLUGIN_HOME", fullDir)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.Setenv("CF_PLUGIN_HOME", old_PLUGINS_HOME)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("exit codes", func() {
		It("exits non-zero when an unknown command is invoked", func() {
			result := Cf("some-command-that-should-never-actually-be-a-real-thing-i-can-use")

			Eventually(result, 3*time.Second).Should(Say("not a registered command"))
			Eventually(result).Should(Exit(1))
		})

		It("exits non-zero when known command is invoked with invalid option", func() {
			result := Cf("push", "--crazy")
			Eventually(result).Should(Exit(1))
		})
	})

	It("can print help for all core commands by executing only the command `cf`", func() {
		output := Cf().Wait(3 * time.Second)
		Eventually(output.Out.Contents).Should(ContainSubstring("A command line tool to interact with Cloud Foundry"))
	})

	Describe("Plugins", func() {
		It("Can call a plugin command from the Plugins configuration if it does not exist as a cf command", func() {
			output := Cf("test_1_cmd1").Wait(3 * time.Second)
			Eventually(output.Out).Should(Say("You called cmd1 in test_1"))
		})

		It("Can call another plugin command when more than one plugin is installed", func() {
			output := Cf("test_2_cmd1").Wait(3 * time.Second)
			Eventually(output.Out).Should(Say("You called cmd1 in test_2"))
		})

		It("informs user for any invalid commands", func() {
			output := Cf("foo-bar")
			Eventually(output.Out, 3*time.Second).Should(Say("'foo-bar' is not a registered command"))
		})

		It("Calls help if the plugin shares the same name", func() {
			output := Cf("help")
			Consistently(output.Out, 1).ShouldNot(Say("You called help in test_with_help"))
		})

		It("Calls the core push command if the plugin shares the same name", func() {
			output := Cf("push")
			Consistently(output.Out, 1).ShouldNot(Say("You called push in test_with_push"))
		})

		It("Calls the core short name if a plugin shares the same name", func() {
			output := Cf("p")
			Consistently(output.Out, 1).ShouldNot(Say("You called p within the plugin"))
		})

		It("Passes all arguments to a plugin", func() {
			output := Cf("my-say", "foo").Wait(3 * time.Second)
			Eventually(output.Out).Should(Say("foo"))
		})

		It("Passes all arguments and flags to a plugin", func() {
			output := Cf("my-say", "foo", "--loud").Wait(3 * time.Second)
			Eventually(output.Out).Should(Say("FOO"))
		})

		It("Calls a plugin that calls core commands", func() {
			output := Cf("awesomeness").Wait(3 * time.Second)
			Eventually(output.Out).Should(Say("my-say")) //look for another plugin
		})

		It("Sends stdoutput to the plugin to echo", func() {
			output := Cf("core-command", "plugins").Wait(3 * time.Second)
			Eventually(output.Out.Contents).Should(MatchRegexp("Command output from the plugin(.*\\W)*awesomeness(.*\\W)*FIN"))
		})

		It("Can call a plugin that requires stdin (interactive)", func() {
			session := CfWithIo("input", "silly\n").Wait(5 * time.Second)
			Eventually(session.Out).Should(Say("silly"))
		})

		It("exits 1 when a plugin panics", func() {
			session := Cf("panic").Wait(5 * time.Second)
			Eventually(session).Should(Exit(1))
		})

		It("exits 1 when a plugin exits 1", func() {
			session := Cf("exit1").Wait(5 * time.Second)
			Eventually(session).Should(Exit(1))
		})
	})
})

func Cf(args ...string) *Session {
	path, err := Build("github.com/cloudfoundry/cli/main")
	Expect(err).NotTo(HaveOccurred())

	session, err := Start(exec.Command(path, args...), GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}
func CfWithIo(command string, args string) *Session {
	path, err := Build("github.com/cloudfoundry/cli/main")
	Expect(err).NotTo(HaveOccurred())

	cmd := exec.Command(path, command)

	stdin, err := cmd.StdinPipe()
	Expect(err).ToNot(HaveOccurred())

	buffer := bufio.NewWriter(stdin)
	buffer.WriteString(args)
	buffer.Flush()

	session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}

// gexec.Build leaves a compiled binary behind in /tmp.
var _ = AfterSuite(func() {
	CleanupBuildArtifacts()
})
