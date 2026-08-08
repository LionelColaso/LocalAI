// run-windows is the run.exe entrypoint shipped inside native windows/amd64
// backend images.
//
// On Linux and macOS a backend image ships run.sh, which picks the best binary
// for the host and sets up its library search path. A Windows image has no
// shell to run a .sh through, so pkg/model/process.go starts run.exe instead
// when it exists next to run.sh. This launcher mirrors the run.sh selection
// logic (see ../run.sh) for Windows: prefer the ggml CPU_ALL_VARIANTS build,
// fall back to the gRPC-RPC build when LLAMACPP_GRPC_SERVERS is set, and only
// then to the static fallback binary.
//
// It is a normal Windows PE built at CI time by scripts/build/llama-cpp-windows.sh
// from this directory; no source change here requires the C++ toolchain.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	exe, err := os.Executable()
	if err != nil {
		os.Exit(1)
	}
	dir := filepath.Dir(exe)

	binary := "llama-cpp-fallback.exe"
	if _, err := os.Stat(filepath.Join(dir, "llama-cpp-cpu-all.exe")); err == nil {
		binary = "llama-cpp-cpu-all.exe"
	}
	if os.Getenv("LLAMACPP_GRPC_SERVERS") != "" {
		if _, err := os.Stat(filepath.Join(dir, "llama-cpp-grpc.exe")); err == nil {
			binary = "llama-cpp-grpc.exe"
		}
	}

	// ggml shared backends (libggml-cpu-*.dll) live next to the executable so
	// ggml's own registry finds them. A lib\ dir is still honoured when a
	// variant ships one, mirroring the run.sh LD_LIBRARY_PATH handling.
	if st, err := os.Stat(filepath.Join(dir, "lib")); err == nil && st.IsDir() {
		_ = os.Setenv("PATH", filepath.Join(dir, "lib")+string(os.PathListSeparator)+os.Getenv("PATH"))
	}

	cmd := exec.Command(filepath.Join(dir, binary), os.Args[1:]...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		os.Exit(1)
	}
}
