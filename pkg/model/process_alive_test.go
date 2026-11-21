package model

import (
	"os/exec"
	"runtime"
	"time"

	process "github.com/mudler/go-processmanager"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProcessAlive", func() {
	It("follows a process from start to stop", func() {
		if runtime.GOOS == "windows" {
			Skip("uses a POSIX sleep binary")
		}
		sleep, err := exec.LookPath("sleep")
		if err != nil {
			Skip("sleep not found")
		}

		notStarted := process.New(process.WithTemporaryStateDir(), process.WithName(sleep))
		Expect(ProcessAlive(notStarted)).To(BeFalse())

		p := process.New(process.WithTemporaryStateDir(), process.WithName(sleep), process.WithArgs("30"))
		Expect(p.Run()).To(Succeed())
		Expect(ProcessAlive(p)).To(BeTrue())
		Expect(p.Stop()).To(Succeed())
		Eventually(p.Done()).WithTimeout(5 * time.Second).Should(BeClosed())
		Expect(ProcessAlive(p)).To(BeFalse())
		Expect(ProcessAlive(nil)).To(BeFalse())
	})
})
