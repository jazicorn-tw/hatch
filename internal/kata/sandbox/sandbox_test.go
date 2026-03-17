package sandbox

// Internal tests — package sandbox — so we can reach unexported helpers
// extractGoPackage and isPass directly.

import (
	"context"
	"strings"
	"testing"

	"github.com/jazicorn/hatch/internal/kata"
)

// ---------------------------------------------------------------------------
// extractGoPackage
// ---------------------------------------------------------------------------

func TestExtractGoPackageMain(t *testing.T) {
	src := "package main\n\nfunc main() {}\n"
	if got := extractGoPackage(src); got != "main" {
		t.Errorf("want main, got %q", got)
	}
}

func TestExtractGoPackageFoo(t *testing.T) {
	src := "// comment\npackage foo\n"
	if got := extractGoPackage(src); got != "foo" {
		t.Errorf("want foo, got %q", got)
	}
}

func TestExtractGoPackageNone(t *testing.T) {
	if got := extractGoPackage("// no package here"); got != "" {
		t.Errorf("want empty, got %q", got)
	}
}

func TestExtractGoPackageEmpty(t *testing.T) {
	if got := extractGoPackage(""); got != "" {
		t.Errorf("want empty for empty input, got %q", got)
	}
}

func TestExtractGoPackageWithLeadingSpaces(t *testing.T) {
	src := "  package mypackage\n"
	if got := extractGoPackage(src); got != "mypackage" {
		t.Errorf("want mypackage, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// isPass
// ---------------------------------------------------------------------------

func TestIsPassGoOkPrefix(t *testing.T) {
	if !isPass("ok  github.com/foo/bar\t0.001s\n") {
		t.Error("expected pass for 'ok ' prefix")
	}
}

func TestIsPassGoOkNewline(t *testing.T) {
	if !isPass("=== RUN   TestFoo\n--- PASS: TestFoo (0.00s)\nok  pkg\n") {
		t.Error("expected pass for '\\nok '")
	}
}

func TestIsPassGoPASS(t *testing.T) {
	if !isPass("=== RUN   TestFoo\n--- PASS: TestFoo (0.00s)\n") {
		t.Error("expected pass for '--- PASS'")
	}
}

func TestIsPassGoFAIL(t *testing.T) {
	out := "--- PASS: TestA (0.00s)\n--- FAIL: TestB (0.01s)\n"
	if isPass(out) {
		t.Error("expected fail when '--- FAIL' present")
	}
}

func TestIsPassPytestPassed(t *testing.T) {
	if !isPass("2 passed, 0 warnings in 0.12s") {
		t.Error("expected pass for pytest ' passed'")
	}
}

func TestIsPassPytestFailed(t *testing.T) {
	if isPass("1 passed, 1 failed in 0.12s") {
		t.Error("expected fail when pytest ' failed' present")
	}
}

func TestIsPassJestPassed(t *testing.T) {
	if !isPass("Tests: 2 passed, 2 total") {
		t.Error("expected pass for jest 'Tests:.*passed'")
	}
}

func TestIsPassJestFailed(t *testing.T) {
	if isPass("Tests: 1 failed, 1 passed, 2 total") {
		t.Error("expected fail when jest 'failed' present")
	}
}

func TestIsPassJUnit(t *testing.T) {
	if !isPass("OK (2 tests)") {
		t.Error("expected pass for JUnit 'OK ('")
	}
}

func TestIsPassEmpty(t *testing.T) {
	if isPass("") {
		t.Error("expected fail for empty output")
	}
}

func TestIsPassUnrecognised(t *testing.T) {
	if isPass("some random output") {
		t.Error("expected fail for unrecognised output")
	}
}

// ---------------------------------------------------------------------------
// lookPath
// ---------------------------------------------------------------------------

func TestLookPathFound(t *testing.T) {
	// "sh" should be on PATH in CI and developer machines.
	p, err := lookPath("sh")
	if err != nil {
		t.Skipf("sh not found: %v", err)
	}
	if p == "" {
		t.Error("expected non-empty path")
	}
}

func TestLookPathNotFound(t *testing.T) {
	_, err := lookPath("__no_such_binary_xyz__")
	if err == nil {
		t.Error("expected error for non-existent binary")
	}
}

// ---------------------------------------------------------------------------
// Run — default timeout
// ---------------------------------------------------------------------------

func TestRunDefaultTimeout(t *testing.T) {
	// Config with zero timeout should apply DefaultTimeout without panicking.
	// Use a trivially passing Go kata so we don't need Python/Node/Java.
	goBin, err := lookPath("go")
	if err != nil {
		t.Skip("go binary not found; skipping Run test")
	}
	_ = goBin

	k := kata.Kata{
		ID:       "test-kata",
		Language: kata.Go,
		Tests: `package kata_test
import "testing"
func TestTrivial(t *testing.T) {}
`,
	}
	solution := "package kata_test\n"

	result, err := Run(context.Background(), k, solution, Config{Timeout: 0})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Result should not panic; pass/fail depends on the solution.
	_ = result
}

// ---------------------------------------------------------------------------
// Run — Go kata end-to-end
// ---------------------------------------------------------------------------

func TestRunGoKataPass(t *testing.T) {
	_, err := lookPath("go")
	if err != nil {
		t.Skip("go binary not found; skipping")
	}

	k := kata.Kata{
		ID:       "add-kata",
		Language: kata.Go,
		Tests: strings.Join([]string{
			"package kata_test",
			"import \"testing\"",
			"func TestAdd(t *testing.T) {",
			"\tif Add(1,2) != 3 { t.Fatal(\"wrong\") }",
			"}",
		}, "\n"),
	}
	solution := strings.Join([]string{
		"package kata_test",
		"func Add(a, b int) int { return a + b }",
	}, "\n")

	result, err := Run(context.Background(), k, solution, Config{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Passed {
		t.Errorf("expected kata to pass; output:\n%s", result.Output)
	}
}

func TestRunGoKataFail(t *testing.T) {
	_, err := lookPath("go")
	if err != nil {
		t.Skip("go binary not found; skipping")
	}

	k := kata.Kata{
		ID:       "fail-kata",
		Language: kata.Go,
		Tests: strings.Join([]string{
			"package kata_test",
			"import \"testing\"",
			"func TestFail(t *testing.T) { t.Fatal(\"always fail\") }",
		}, "\n"),
	}

	result, err := Run(context.Background(), k, "package kata_test\n", Config{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Passed {
		t.Error("expected kata to fail")
	}
}

// ---------------------------------------------------------------------------
// Run — language dispatch (Python / JavaScript / Java)
// These tests exercise the switch cases in run(). If the required binary is
// absent the runner returns an error early (lookPath fails), so the test
// accepts either outcome.
// ---------------------------------------------------------------------------

func TestRunPythonKata(t *testing.T) {
	k := kata.Kata{
		ID:       "py-kata",
		Language: kata.Python,
		Tests:    "def test_trivial(): pass\n",
	}
	// Run must not panic regardless of whether python3 is installed.
	result, err := Run(context.Background(), k, "x = 1\n", Config{})
	if err != nil {
		t.Fatalf("Run (Python): unexpected Go-level error: %v", err)
	}
	// pass/fail depends on whether python3/pytest is available; we just check no panic.
	_ = result.Passed
}

func TestRunJavaScriptKata(t *testing.T) {
	k := kata.Kata{
		ID:       "js-kata",
		Language: kata.JavaScript,
		Tests:    "test('trivial', () => {});\n",
	}
	result, err := Run(context.Background(), k, "// solution\n", Config{})
	if err != nil {
		t.Fatalf("Run (JavaScript): unexpected Go-level error: %v", err)
	}
	_ = result.Passed
}

func TestRunJavaKata(t *testing.T) {
	k := kata.Kata{
		ID:       "java-kata",
		Language: kata.Java,
		Tests:    "public class KataTest {}\n",
	}
	result, err := Run(context.Background(), k, "public class Kata {}\n", Config{})
	if err != nil {
		t.Fatalf("Run (Java): unexpected Go-level error: %v", err)
	}
	_ = result.Passed
}
