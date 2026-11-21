//go:build !windows

package fsperm

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// mode returns the permission bits of path.
func mode(path string) os.FileMode {
	GinkgoHelper()
	st, err := os.Stat(path)
	Expect(err).ToNot(HaveOccurred())
	return st.Mode().Perm()
}

var _ = Describe("fsperm on Unix", func() {
	It("removes the group and other bits", func() {
		dir := GinkgoT().TempDir()
		file := filepath.Join(dir, "f")
		Expect(os.WriteFile(file, []byte("x"), 0o644)).To(Succeed())
		Expect(os.Chmod(file, 0o755)).To(Succeed())
		Expect(Restrict(file)).To(Succeed())
		Expect(mode(file)).To(Equal(os.FileMode(0o700)))

		sub := filepath.Join(dir, "d")
		Expect(os.Mkdir(sub, 0o755)).To(Succeed())
		Expect(os.Chmod(sub, 0o755)).To(Succeed())
		Expect(Restrict(sub)).To(Succeed())
		Expect(mode(sub)).To(Equal(os.FileMode(0o700)))

		err := Restrict(filepath.Join(dir, "missing"))
		Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
	})

	It("checks whether a file is private", func() {
		dir := GinkgoT().TempDir()
		file := filepath.Join(dir, "key")
		Expect(os.WriteFile(file, []byte("x"), 0o600)).To(Succeed())
		Expect(Check(file)).To(Succeed())

		Expect(os.Chmod(file, 0o640)).To(Succeed())
		err := Check(file)
		Expect(errors.Is(err, ErrNotPrivate)).To(BeTrue())

		err = Check(filepath.Join(dir, "missing"))
		Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
	})

	It("keeps the mode on Unix when the file is not private", func() {
		file := filepath.Join(GinkgoT().TempDir(), "key")
		Expect(os.WriteFile(file, []byte("x"), 0o600)).To(Succeed())
		Expect(os.Chmod(file, 0o644)).To(Succeed())

		restricted, err := CheckOrRestrict(file)
		Expect(restricted).To(BeFalse())
		Expect(errors.Is(err, ErrNotPrivate)).To(BeTrue())
		Expect(mode(file)).To(Equal(os.FileMode(0o644)))

		Expect(os.Chmod(file, 0o600)).To(Succeed())
		restricted, err = CheckOrRestrict(file)
		Expect(restricted).To(BeFalse())
		Expect(err).ToNot(HaveOccurred())
	})

	It("writes files and creates private folders", func() {
		dir := GinkgoT().TempDir()
		priv := filepath.Join(dir, "a", "b")
		Expect(os.MkdirAll(priv, 0o755)).To(Succeed())
		Expect(os.Chmod(priv, 0o755)).To(Succeed())

		// MkdirAll closes a folder that already exists.
		Expect(MkdirAll(priv)).To(Succeed())
		Expect(mode(priv)).To(Equal(os.FileMode(0o700)))

		secret := filepath.Join(priv, "secret")
		Expect(WriteFile(secret, []byte("s"), 0o600)).To(Succeed())
		Expect(mode(secret)).To(Equal(os.FileMode(0o600)))

		public := filepath.Join(priv, "public")
		Expect(WriteFile(public, []byte("p"), 0o644)).To(Succeed())
		Expect(os.Chmod(public, 0o644)).To(Succeed())
		Expect(WriteFile(public, []byte("p"), 0o644)).To(Succeed())
		Expect(mode(public)).To(Equal(os.FileMode(0o644)))

		err := WriteFile(filepath.Join(dir, "no", "such", "dir"), nil, 0o600)
		Expect(err).To(HaveOccurred())
	})

	DescribeTable("follows the executable rules of the platform",
		func(goos, path string, perm os.FileMode, want bool) {
			Expect(isExecutable(goos, path, perm)).To(Equal(want))
		},
		Entry("a Linux shell script", "linux", "run.sh", os.FileMode(0o755), true),
		Entry("a Linux shell script without the group bit", "linux", "run.sh", os.FileMode(0o100), true),
		Entry("a Linux file without the executable bit", "linux", "run.sh", os.FileMode(0o644), false),
		Entry("a Windows executable on Linux", "linux", "llama.exe", os.FileMode(0o644), false),
		Entry("a Linux folder", "linux", "bin", os.FileMode(os.ModeDir|0o755), false),
		Entry("a Windows executable", "windows", `C:\b\llama-cpp.EXE`, os.FileMode(0o666), true),
		Entry("a PowerShell script", "windows", "run.ps1", os.FileMode(0o666), true),
		Entry("a command script", "windows", "run.cmd", os.FileMode(0o444), true),
		Entry("a batch script", "windows", "run.bat", os.FileMode(0o666), true),
		Entry("a shell script on Windows", "windows", "run.sh", os.FileMode(0o777), false),
		Entry("a model file on Windows", "windows", "model.gguf", os.FileMode(0o666), false),
		Entry("a file without an extension on Windows", "windows", "noext", os.FileMode(0o777), false),
		Entry("a folder with an executable extension on Windows", "windows", "dir.exe", os.FileMode(os.ModeDir|0o777), false),
	)

	It("follows the Unix rule for the exported helper", func() {
		Expect(IsExecutable("x", 0o700)).To(BeTrue())
		Expect(IsExecutable("x.exe", 0o600)).To(BeFalse())
	})

	It("checks the mode the caller holds", func() {
		file := filepath.Join(GinkgoT().TempDir(), "key")
		Expect(os.WriteFile(file, []byte("x"), 0o600)).To(Succeed())
		Expect(os.Chmod(file, 0o644)).To(Succeed())

		info, err := os.Stat(file)
		Expect(err).ToNot(HaveOccurred())

		restricted, err := RequirePrivate(file, info)
		Expect(restricted).To(BeFalse())
		Expect(errors.Is(err, ErrNotPrivate)).To(BeTrue())
		Expect(mode(file)).To(Equal(os.FileMode(0o644)))

		// The check uses the mode the caller read, not the path, so a file that
		// changes after the caller opened it does not change the answer.
		Expect(os.Chmod(file, 0o600)).To(Succeed())

		_, err = RequirePrivate(file, info)
		Expect(errors.Is(err, ErrNotPrivate)).To(BeTrue())

		info, err = os.Stat(file)
		Expect(err).ToNot(HaveOccurred())

		restricted, err = RequirePrivate(file, info)
		Expect(restricted).To(BeFalse())
		Expect(err).ToNot(HaveOccurred())
	})
})
