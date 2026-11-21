//go:build windows

package model

import (
	"os/exec"
	"time"

	process "github.com/mudler/go-processmanager"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The Windows spec checks liveness and Stop through the process manager, with
// the same job object option that startProcess uses.
var _ = Describe("ProcessAlive on Windows", func() {
	It("follows a process from start to stop", func() {
		powershell, err := exec.LookPath("powershell.exe")
		if err != nil {
			Skip("powershell.exe not found")
		}
		p := process.New(
			process.WithTemporaryStateDir(),
			process.WithName(powershell),
			process.WithArgs("-NoProfile", "-Command", "Start-Sleep -Seconds 60"),
			process.WithJobObject(true),
		)
		Expect(p.Run()).To(Succeed())
		Expect(ProcessAlive(p)).To(BeTrue())
		Expect(p.Stop()).To(Succeed())
		Eventually(p.Done()).WithTimeout(10 * time.Second).Should(BeClosed())
		Expect(ProcessAlive(p)).To(BeFalse())
	})
})
