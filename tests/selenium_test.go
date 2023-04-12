package tests

import (
	"fmt"
	"net"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/go-auxiliaries/selenium/tests/testconf"

	"github.com/blang/semver"
	"github.com/go-auxiliaries/selenium"
	"github.com/go-auxiliaries/selenium/internal/seleniumtest"
)

func pickUnusedPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	if err := l.Close(); err != nil {
		return 0, err
	}
	return port, nil
}

func TestChrome(t *testing.T) {
	if testconf.UseDocker() {
		t.Skip("Skipping Chrome tests because they will be run under a Docker container")
	}
	if testconf.ChromeBinary() == "" {
		t.Skip("Skipping Chrome tests because binary is not available")
	}
	if testconf.ChromeDriverPath() == "" {
		t.Skipf("Skipping Chrome tests because ChromeDriver is not available")
	}

	t.Run("Chromedriver", func(t *testing.T) {
		runChromeTests(t, seleniumtest.Config{
			Path: testconf.ChromeBinary(),
		})
	})

	t.Run("Selenium3", func(t *testing.T) {
		runChromeTests(t, seleniumtest.Config{
			Path:            testconf.ChromeBinary(),
			SeleniumVersion: semver.MustParse("3.0.0"),
		})
	})
}

func runChromeTests(t *testing.T, c seleniumtest.Config) {
	c.Browser = "chrome"
	c.Headless = testconf.Headless()

	if testconf.StartFrameBuffer() {
		c.ServiceOptions = append(c.ServiceOptions, selenium.StartFrameBuffer())
	}
	if testing.Verbose() {
		selenium.SetDebug(true)
		c.ServiceOptions = append(c.ServiceOptions, selenium.Output(os.Stderr))
	}

	port, err := pickUnusedPort()
	if err != nil {
		t.Fatalf("pickUnusedPort() returned error: %v", err)
	}
	c.Addr = fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port)

	var s *selenium.Service
	if c.SeleniumVersion.Major == 3 {
		c.ServiceOptions = append(c.ServiceOptions, selenium.ChromeDriver(testconf.ChromeDriverPath()))
		s, err = selenium.NewSeleniumService(testconf.Selenium3Path(), port, c.ServiceOptions...)
	} else {
		s, err = selenium.NewChromeDriverService(testconf.ChromeDriverPath(), port, c.ServiceOptions...)
	}
	if err != nil {
		t.Fatalf("Error starting the server: %v", err)
	}

	hs := httptest.NewServer(seleniumtest.Handler)
	defer hs.Close()
	c.ServerURL = hs.URL

	seleniumtest.RunCommonTests(t, c)
	seleniumtest.RunChromeTests(t, c)

	if err := s.Stop(); err != nil {
		t.Fatalf("Error stopping the ChromeDriver service: %v", err)
	}
}

func TestFirefox(t *testing.T) {
	if testconf.UseDocker() {
		t.Skip("Skipping tests because they will be run under a Docker container")
	}
	if _, err := os.Stat(testconf.GeckoDriverPath()); err != nil {
		t.Skipf("Skipping Firefox tests on Selenium 3 because geckodriver binary %q not found", testconf.GeckoDriverPath())
	}

	if testconf.FirefoxBinarySelenium3() == "" {
		t.Skipf("Skipping Firefox tests because binary is not available")
	}
	t.Run("Selenium3", func(t *testing.T) {
		runFirefoxTests(t, testconf.Selenium3Path(), seleniumtest.Config{
			SeleniumVersion: semver.MustParse("3.0.0"),
			ServiceOptions:  []selenium.ServiceOption{selenium.GeckoDriver(testconf.GeckoDriverPath())},
			Path:            testconf.FirefoxBinarySelenium3(),
		})
	})
	t.Run("Geckodriver", func(t *testing.T) {
		runFirefoxTests(t, testconf.GeckoDriverPath(), seleniumtest.Config{
			Path: testconf.FirefoxBinarySelenium3(),
		})
	})
}

func TestHTMLUnit(t *testing.T) {
	if testconf.UseDocker() {
		t.Skip("Skipping tests because they will be run under a Docker container")
	}
	if _, err := os.Stat(testconf.Selenium3Path()); err != nil {
		t.Skipf("Skipping HTMLUnit tests because the Selenium WebDriver JAR was not found at path %q", testconf.Selenium3Path())
	}
	if _, err := os.Stat(testconf.HtmlUnitDriverPath()); err != nil {
		t.Skipf("Skipping HTMLUnit tests because the HTMLUnit Driver JAR not found at path %q", testconf.HtmlUnitDriverPath())
	}

	if testing.Verbose() {
		selenium.SetDebug(true)
	}

	c := seleniumtest.Config{
		Browser:         "htmlunit",
		SeleniumVersion: semver.MustParse("3.0.0"),
		ServiceOptions:  []selenium.ServiceOption{selenium.HTMLUnit(testconf.HtmlUnitDriverPath())},
		// HTMLUnit-Driver currently does not support the sameSite attribute
		// See: https://github.com/SeleniumHQ/htmlunit-driver/issues/97
		SameSiteUnsupported: true,
	}

	port, err := pickUnusedPort()
	if err != nil {
		t.Fatalf("pickUnusedPort() returned error: %v", err)
	}
	s, err := selenium.NewSeleniumService(testconf.Selenium3Path(), port, c.ServiceOptions...)
	if err != nil {
		t.Fatalf("Error starting the WebDriver server with binary %q: %v", testconf.Selenium3Path(), err)
	}
	c.Addr = fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port)

	hs := httptest.NewServer(seleniumtest.Handler)
	defer hs.Close()
	c.ServerURL = hs.URL

	seleniumtest.RunCommonTests(t, c)

	if err := s.Stop(); err != nil {
		t.Fatalf("Error stopping the Selenium service: %v", err)
	}
}

func runFirefoxTests(t *testing.T, webDriverPath string, c seleniumtest.Config) {
	c.Browser = "firefox"

	if testconf.StartFrameBuffer() {
		c.ServiceOptions = append(c.ServiceOptions, selenium.StartFrameBuffer())
	}
	if testing.Verbose() {
		selenium.SetDebug(true)
		c.ServiceOptions = append(c.ServiceOptions, selenium.Output(os.Stderr))
	}
	if testconf.JavaPath() != "" {
		c.ServiceOptions = append(c.ServiceOptions, selenium.JavaPath(testconf.JavaPath()))
	}

	port, err := pickUnusedPort()
	if err != nil {
		t.Fatalf("pickUnusedPort() returned error: %v", err)
	}

	var s *selenium.Service
	if c.SeleniumVersion.Major == 0 {
		c.Addr = fmt.Sprintf("http://127.0.0.1:%d", port)
		s, err = selenium.NewGeckoDriverService(webDriverPath, port, c.ServiceOptions...)
	} else {
		c.Addr = fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port)
		if _, err := os.Stat(testconf.Selenium3Path()); err != nil {
			t.Skipf("Skipping Firefox tests using Selenium 3 because Selenium WebDriver JAR not found at path %q", testconf.Selenium3Path())
		}

		s, err = selenium.NewSeleniumService(webDriverPath, port, c.ServiceOptions...)
	}
	if err != nil {
		t.Fatalf("Error starting the WebDriver server with binary %q: %v", webDriverPath, err)
	}

	hs := httptest.NewServer(seleniumtest.Handler)
	defer hs.Close()
	c.ServerURL = hs.URL

	if c.SeleniumVersion.Major == 0 {
		c.Addr = fmt.Sprintf("http://127.0.0.1:%d", port)
	} else {
		c.Addr = fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port)
	}

	c.Headless = testconf.Headless()

	seleniumtest.RunCommonTests(t, c)
	seleniumtest.RunFirefoxTests(t, c)

	if err := s.Stop(); err != nil {
		t.Fatalf("Error stopping the Selenium service: %v", err)
	}
}

func TestDocker(t *testing.T) {
	if testconf.RunningUnderDocker() {
		return
	}
	if !testconf.UseDocker() {
		t.Skip("Skipping Docker tests because --docker was not specified.")
	}

	args := []string{"build", "-t", "go-selenium", "testing/"}
	if out, err := exec.Command("docker", args...).CombinedOutput(); err != nil {
		t.Logf("Output from `docker %s`:\n%s", strings.Join(args, " "), string(out))
		t.Fatalf("Building Docker container failed: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() returned error: %v", err)
	}

	// TODO(minusnine): pass through relevant flags to docker-test.sh to be
	// passed to go test.
	cmd := exec.Command("docker", "run", fmt.Sprintf("--volume=%s:/code", cwd), "--workdir=/code/", "go-selenium", "testing/docker-test.sh")
	if testing.Verbose() {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		t.Fatalf("docker run failed: %v", err)
	}
}
