//go:build windows

package fsperm

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sys/windows"
)

// grantEveryone adds an entry that lets every user read path, the way an
// inherited list of a shared folder does.
func grantEveryone(path string) {
	GinkgoHelper()
	everyone, err := windows.CreateWellKnownSid(windows.WinWorldSid)
	Expect(err).ToNot(HaveOccurred())
	sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	Expect(err).ToNot(HaveOccurred())
	old, _, err := sd.DACL()
	Expect(err).ToNot(HaveOccurred())
	acl, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.GENERIC_READ,
		AccessMode:        windows.GRANT_ACCESS,
		Inheritance:       windows.NO_INHERITANCE,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_WELL_KNOWN_GROUP,
			TrusteeValue: windows.TrusteeValueFromSID(everyone),
		},
	}}, old)
	Expect(err).ToNot(HaveOccurred())
	Expect(windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION, nil, nil, acl, nil)).To(Succeed())
}

var _ = Describe("fsperm on Windows", func() {
	It("restricts a file and checks it", func() {
		file := filepath.Join(GinkgoT().TempDir(), "key")
		Expect(os.WriteFile(file, []byte("x"), 0o600)).To(Succeed())
		grantEveryone(file)

		err := Check(file)
		Expect(errors.Is(err, ErrNotPrivate)).To(BeTrue())

		Expect(Restrict(file)).To(Succeed())
		Expect(Check(file)).To(Succeed())

		b, err := os.ReadFile(file)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(b)).To(Equal("x"))
	})

	It("repairs a file that other users can read", func() {
		file := filepath.Join(GinkgoT().TempDir(), "key")
		Expect(os.WriteFile(file, []byte("x"), 0o600)).To(Succeed())
		grantEveryone(file)

		restricted, err := CheckOrRestrict(file)
		Expect(err).ToNot(HaveOccurred())
		Expect(restricted).To(BeTrue())
		Expect(Check(file)).To(Succeed())
	})

	It("creates a private folder and writes a file", func() {
		dir := filepath.Join(GinkgoT().TempDir(), "state")
		Expect(MkdirAll(dir)).To(Succeed())
		Expect(Check(dir)).To(Succeed())

		file := filepath.Join(dir, "secret")
		Expect(WriteFile(file, []byte("s"), 0o600)).To(Succeed())
		Expect(Check(file)).To(Succeed())
	})

	It("follows the extension rule for executables", func() {
		Expect(IsExecutable(`C:\x\llama.exe`, 0o666)).To(BeTrue())
		Expect(IsExecutable(`C:\x\run.sh`, 0o777)).To(BeFalse())
	})

	It("repairs a file that needs to be private", func() {
		file := filepath.Join(GinkgoT().TempDir(), "key")
		Expect(os.WriteFile(file, []byte("x"), 0o600)).To(Succeed())
		grantEveryone(file)

		info, err := os.Stat(file)
		Expect(err).ToNot(HaveOccurred())

		restricted, err := RequirePrivate(file, info)
		Expect(err).ToNot(HaveOccurred())
		Expect(restricted).To(BeTrue())
		Expect(Check(file)).To(Succeed())

		restricted, err = RequirePrivate(file, info)
		Expect(err).ToNot(HaveOccurred())
		Expect(restricted).To(BeFalse())
	})
})
