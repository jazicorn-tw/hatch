// Package sandbox runs a user's kata solution against the kata's test cases
// in an isolated temporary directory using a subprocess with a configurable timeout.
package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jazicorn/hatch/internal/kata"
)

// DefaultTimeout is the default execution timeout for kata runs.
const DefaultTimeout = 30 * time.Second

const (
	errWriteSolution = "sandbox: write solution: %w"
	errWriteTests    = "sandbox: write tests: %w"
)

// Config configures sandbox execution.
type Config struct {
	// Timeout is the maximum time allowed for the subprocess. Defaults to DefaultTimeout.
	Timeout time.Duration
}

// Result holds the outcome of a sandbox run.
type Result struct {
	Passed   bool
	Output   string
	Duration time.Duration
}

// Run writes the user's solution and the kata's tests to a temp directory and
// executes them in a subprocess. It returns the pass/fail result and output.
func Run(ctx context.Context, k kata.Kata, solution string, cfg Config) (Result, error) {
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeout
	}

	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("kata-%s-", k.ID))
	if err != nil {
		return Result{}, fmt.Errorf("sandbox: create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	start := time.Now()
	out, err := run(ctx, k, solution, tmpDir, cfg.Timeout)
	elapsed := time.Since(start)

	passed := err == nil && isPass(string(out))
	return Result{
		Passed:   passed,
		Output:   string(out),
		Duration: elapsed,
	}, nil
}

// run dispatches to the language-specific runner.
func run(ctx context.Context, k kata.Kata, solution, tmpDir string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	switch k.Language {
	case kata.Python:
		return runPython(ctx, k, solution, tmpDir)
	case kata.JavaScript:
		return runJavaScript(ctx, k, solution, tmpDir)
	case kata.Java:
		return runJava(ctx, k, solution, tmpDir)
	default: // kata.Go and fallback
		return runGo(ctx, k, solution, tmpDir)
	}
}

// lookPath resolves a command name to its absolute path, returning an error if not found.
func lookPath(name string) (string, error) {
	p, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("sandbox: %q not found in PATH: %w", name, err)
	}
	return p, nil
}

// ---- Go ----

func runGo(ctx context.Context, k kata.Kata, solution, tmpDir string) ([]byte, error) {
	goBin, err := lookPath("go")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "kata.go"), []byte(solution), 0600); err != nil {
		return nil, fmt.Errorf(errWriteSolution, err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "kata_test.go"), []byte(k.Tests), 0600); err != nil {
		return nil, fmt.Errorf(errWriteTests, err)
	}

	// Minimal go.mod so "go test" doesn't try to fetch a module.
	pkg := "kata"
	if p := extractGoPackage(solution); p != "" {
		pkg = p
	}
	goMod := fmt.Sprintf("module %s\n\ngo 1.21\n", pkg)
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0600); err != nil {
		return nil, fmt.Errorf("sandbox: write go.mod: %w", err)
	}

	cmd := exec.CommandContext(ctx, goBin, "test", "-v", "./...")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(),
		"GOPROXY=off",
		"GONOSUMCHECK=*",
		"GOFLAGS=",
	)
	return cmd.CombinedOutput()
}

// extractGoPackage returns the package name declared in Go source, or "".
func extractGoPackage(src string) string {
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "package "))
		}
	}
	return ""
}

// ---- Python ----

func runPython(ctx context.Context, k kata.Kata, solution, tmpDir string) ([]byte, error) {
	python, err := lookPath("python3")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "kata.py"), []byte(solution), 0600); err != nil {
		return nil, fmt.Errorf(errWriteSolution, err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "test_kata.py"), []byte(k.Tests), 0600); err != nil {
		return nil, fmt.Errorf(errWriteTests, err)
	}

	cmd := exec.CommandContext(ctx, python, "-m", "pytest", "test_kata.py", "-v", "--tb=short")
	cmd.Dir = tmpDir
	return cmd.CombinedOutput()
}

// ---- JavaScript ----

func runJavaScript(ctx context.Context, k kata.Kata, solution, tmpDir string) ([]byte, error) {
	npx, err := lookPath("npx")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "kata.js"), []byte(solution), 0600); err != nil {
		return nil, fmt.Errorf(errWriteSolution, err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "kata.test.js"), []byte(k.Tests), 0600); err != nil {
		return nil, fmt.Errorf(errWriteTests, err)
	}

	// Minimal package.json so node can find the test runner.
	pkgJSON := `{"name":"kata","version":"1.0.0","scripts":{"test":"node --experimental-vm-modules node_modules/.bin/jest"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0600); err != nil {
		return nil, fmt.Errorf("sandbox: write package.json: %w", err)
	}

	cmd := exec.CommandContext(ctx, npx, "--yes", "jest", "--no-coverage", "kata.test.js")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "CI=true")
	return cmd.CombinedOutput()
}

// ---- Java ----

func runJava(ctx context.Context, k kata.Kata, solution, tmpDir string) ([]byte, error) {
	javac, err := lookPath("javac")
	if err != nil {
		return nil, err
	}
	java, err := lookPath("java")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "Kata.java"), []byte(solution), 0600); err != nil {
		return nil, fmt.Errorf(errWriteSolution, err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "KataTest.java"), []byte(k.Tests), 0600); err != nil {
		return nil, fmt.Errorf(errWriteTests, err)
	}

	// Compile.
	compile := exec.CommandContext(ctx, javac, "Kata.java", "KataTest.java")
	compile.Dir = tmpDir
	if out, err := compile.CombinedOutput(); err != nil {
		return out, fmt.Errorf("sandbox: javac: %w", err)
	}

	// Run with JUnit 4 console launcher (assumes junit-platform-console-standalone on PATH or CLASSPATH).
	runCmd := exec.CommandContext(ctx, java, "-cp", ".", "org.junit.runner.JUnitCore", "KataTest")
	runCmd.Dir = tmpDir
	return runCmd.CombinedOutput()
}

// isPass returns true when the test output indicates all tests passed.
// Each runner uses its own conventions.
func isPass(output string) bool {
	// Go: "ok" prefix or PASS
	if strings.Contains(output, "\nok ") || strings.HasPrefix(output, "ok ") {
		return true
	}
	if strings.Contains(output, "--- PASS") && !strings.Contains(output, "--- FAIL") {
		return true
	}
	// pytest: "passed" in summary, no "failed"
	if strings.Contains(output, " passed") && !strings.Contains(output, " failed") {
		return true
	}
	// jest: "Tests:.*passed"
	if strings.Contains(output, "Tests:") && strings.Contains(output, "passed") && !strings.Contains(output, "failed") {
		return true
	}
	// JUnit: "OK ("
	if strings.Contains(output, "OK (") {
		return true
	}
	return false
}
