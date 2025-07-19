package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	fmt.Println("🚀 Starting Build and Verification Test")
	fmt.Println("======================================")

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Go Module Verification", testGoMod},
		{"Code Compilation", testBuild},
		{"Unit Tests", testUnitTests},
		{"Code Formatting", testGoFmt},
		{"Go Vet", testGoVet},
		{"Docker Configuration", testDockerConfig},
		{"API Documentation", testSwaggerDocs},
		{"Environment Configuration", testEnvConfig},
		{"Project Structure", testProjectStructure},
	}

	passed := 0
	failed := 0

	for _, test := range tests {
		fmt.Printf("\n🧪 Running test: %s\n", test.name)
		if err := test.fn(); err != nil {
			fmt.Printf("❌ FAILED: %s - %v\n", test.name, err)
			failed++
		} else {
			fmt.Printf("✅ PASSED: %s\n", test.name)
			passed++
		}
	}

	fmt.Printf("\n📊 Build and Verification Results\n")
	fmt.Printf("=================================\n")
	fmt.Printf("✅ Passed: %d\n", passed)
	fmt.Printf("❌ Failed: %d\n", failed)
	fmt.Printf("📈 Total:  %d\n", passed+failed)

	if failed > 0 {
		fmt.Printf("\n❌ Some tests failed. Please check the output above.\n")
		os.Exit(1)
	} else {
		fmt.Printf("\n🎉 All build and verification tests passed!\n")
		fmt.Printf("The Todo API backend is ready for deployment.\n")
	}
}

func testGoMod() error {
	// Check go.mod exists
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("go.mod file not found")
	}

	// Verify dependencies
	cmd := exec.Command("go", "mod", "verify")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go mod verify failed: %v\nOutput: %s", err, string(output))
	}

	// Check for tidy
	cmd = exec.Command("go", "mod", "tidy")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go mod tidy failed: %v\nOutput: %s", err, string(output))
	}

	fmt.Printf("   Go modules are properly configured\n")
	return nil
}

func testBuild() error {
	// Build the main application
	cmd := exec.Command("go", "build", "-o", "temp_server", "./cmd/server")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %v\nOutput: %s", err, string(output))
	}

	// Clean up
	os.Remove("temp_server")
	os.Remove("temp_server.exe")

	fmt.Printf("   Application builds successfully\n")
	return nil
}

func testUnitTests() error {
	cmd := exec.Command("go", "test", "-v", "./...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unit tests failed: %v\nOutput: %s", err, string(output))
	}

	// Count test results
	outputStr := string(output)
	passCount := strings.Count(outputStr, "PASS:")
	failCount := strings.Count(outputStr, "FAIL:")

	fmt.Printf("   Unit tests: %d passed, %d failed\n", passCount, failCount)
	
	if failCount > 0 {
		return fmt.Errorf("some unit tests failed")
	}

	return nil
}

func testGoFmt() error {
	cmd := exec.Command("gofmt", "-l", ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gofmt check failed: %v", err)
	}

	if len(output) > 0 {
		return fmt.Errorf("code is not formatted. Files need formatting:\n%s", string(output))
	}

	fmt.Printf("   Code is properly formatted\n")
	return nil
}

func testGoVet() error {
	cmd := exec.Command("go", "vet", "./...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go vet failed: %v\nOutput: %s", err, string(output))
	}

	fmt.Printf("   Code passes go vet checks\n")
	return nil
}

func testDockerConfig() error {
	// Check Dockerfile exists
	if _, err := os.Stat("Dockerfile"); os.IsNotExist(err) {
		return fmt.Errorf("Dockerfile not found")
	}

	// Check docker-compose.yml exists
	if _, err := os.Stat("docker-compose.yml"); os.IsNotExist(err) {
		return fmt.Errorf("docker-compose.yml not found")
	}

	// Check docker-compose.test.yml exists
	if _, err := os.Stat("docker-compose.test.yml"); os.IsNotExist(err) {
		return fmt.Errorf("docker-compose.test.yml not found")
	}

	fmt.Printf("   Docker configuration files are present\n")
	return nil
}

func testSwaggerDocs() error {
	// Check swagger files exist
	swaggerFiles := []string{
		"docs/swagger.json",
		"docs/swagger.yaml",
		"docs/docs.go",
	}

	for _, file := range swaggerFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("swagger file not found: %s", file)
		}
	}

	fmt.Printf("   API documentation files are present\n")
	return nil
}

func testEnvConfig() error {
	// Check environment files exist
	envFiles := []string{
		".env.example",
		".env.production",
		".env.test",
		".env.docker",
	}

	for _, file := range envFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("environment file not found: %s", file)
		}
	}

	fmt.Printf("   Environment configuration files are present\n")
	return nil
}

func testProjectStructure() error {
	// Check required directories exist
	requiredDirs := []string{
		"cmd/server",
		"internal/config",
		"internal/handler",
		"internal/middleware",
		"internal/model",
		"internal/repository",
		"internal/service",
		"internal/database",
		"pkg/jwt",
		"pkg/password",
		"pkg/validator",
		"tests/handler",
		"tests/service",
		"tests/repository",
		"tests/middleware",
		"tests/integration",
		"docs",
		"scripts",
	}

	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("required directory not found: %s", dir)
		}
	}

	// Check key files exist
	keyFiles := []string{
		"cmd/server/main.go",
		"go.mod",
		"go.sum",
		"README.md",
		"Makefile",
	}

	for _, file := range keyFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("key file not found: %s", file)
		}
	}

	fmt.Printf("   Project structure is complete\n")
	return nil
}