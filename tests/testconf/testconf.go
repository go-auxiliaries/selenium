package testconf

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/golang/glog"
)

var (
	selenium3Path          = flag.String("selenium3_path", "", "The path to the Selenium 3 server JAR. If empty or the file is not present, Firefox tests using Selenium 3 will not be run.")
	firefoxBinarySelenium3 = flag.String("firefox_binary_for_selenium3", "vendor/firefox/firefox", "The name of the Firefox binary for Selenium 3 tests or the path to it. If the name does not contain directory separators, the PATH will be searched.")
	geckoDriverPath        = flag.String("geckodriver_path", "", "The path to the geckodriver binary. If empty or the file is not present, the Geckodriver tests will not be run.")
	javaPath               = flag.String("java_path", "", "The path to the Java runtime binary to invoke. If not specified, 'java' will be used.")

	chromeDriverPath = flag.String("chrome_driver_path", "", "The path to the ChromeDriver binary. If empty or the file is not present, Chrome tests will not be run.")
	chromeBinary     = flag.String("chrome_binary", "vendor/chrome-linux/chrome", "The name of the Chrome binary or the path to it. If name is not an exact path, the PATH will be searched.")

	htmlUnitDriverPath = flag.String("htmlunit_driver_path", "vendor/htmlunit-driver.jar", "The path to the HTMLUnit Driver JAR.")

	useDocker          = flag.Bool("docker", false, "If set, run the tests in a Docker container.")
	runningUnderDocker = flag.Bool("running_under_docker", false, "This is set by the Docker test harness and should not be needed otherwise.")

	startFrameBuffer = flag.Bool("start_frame_buffer", false, "If true, start an Xvfb subprocess and run the browsers in that X server.")
	headless         = flag.Bool("headless", true, "If true, run Chrome and Firefox in headless mode, not requiring a frame buffer.")

	xvfb = flag.Bool("xvfb", false, "If set, run xvfb tests.")
)

func Selenium3Path() string {
	return strPtrToStrDefault(selenium3Path, func() string {
		return findBestPath("vendor/selenium-server*" /*binary=*/, false)
	})
}

func FirefoxBinarySelenium3() string {
	return strPtrToStr(firefoxBinarySelenium3)
}

func GeckoDriverPath() string {
	return strPtrToStrDefault(geckoDriverPath, func() string {
		return findBestPath("vendor/geckodriver*" /*binary=*/, true)
	})
}

func JavaPath() string {
	return strPtrToStr(javaPath)
}

func ChromeDriverPath() string {
	return strPtrToStrDefault(chromeDriverPath, func() string {
		return findBestPath("vendor/geckodriver*" /*binary=*/, true)
	})
}

func ChromeBinary() string {
	return strPtrToStr(chromeBinary)
}

func HtmlUnitDriverPath() string {
	return strPtrToStr(htmlUnitDriverPath)
}

func UseDocker() bool {
	return boolPtrToBool(useDocker)
}

func RunningUnderDocker() bool {
	return boolPtrToBool(runningUnderDocker)
}

func StartFrameBuffer() bool {
	return boolPtrToBool(startFrameBuffer)
}

func Headless() bool {
	return boolPtrToBool(headless)
}

func Xvfb() bool {
	return boolPtrToBool(xvfb)
}

func findBestPath(glob string, binary bool) string {
	matches, err := filepath.Glob(glob)
	if err != nil {
		glog.Warningf("Error globbing %q: %s", glob, err)
		return ""
	}
	if len(matches) == 0 {
		return ""
	}
	// Iterate backwards: newer versions should be sorted to the end.
	sort.Strings(matches)
	for i := len(matches) - 1; i >= 0; i-- {
		path := matches[i]
		fi, err := os.Stat(path)
		if err != nil {
			glog.Warningf("Error statting %q: %s", path, err)
			continue
		}
		if !fi.Mode().IsRegular() {
			continue
		}
		if binary && fi.Mode().Perm()&0111 == 0 {
			continue
		}
		return path
	}
	return ""
}

func strPtrToStr(val *string) string {
	parseFlags()
	if val == nil {
		return ""
	}
	return *val
}

func strPtrToStrDefault(val *string, def func() string) string {
	parseFlags()
	if val == nil || *val == "" {
		tmp := def()
		val = &tmp
	}
	return *val
}

func boolPtrToBool(val *bool) bool {
	parseFlags()
	if val == nil {
		return false
	}
	return *val
}

func lookupPath(path string) string {
	if path == "" {
		return ""
	}
	if _, err := os.Stat(path); err != nil {
		path, _ = exec.LookPath(path)
	}
	return path
}

func parseFlags() {
	if flag.Parsed() {
		return
	}
	flag.Parse()
	resolveChromeBinary()
	resolveFirefoxBinarySelenium3()
}

func resolveChromeBinary() {
	path := lookupPath(ChromeBinary())
	chromeBinary = &path
}

func resolveFirefoxBinarySelenium3() {
	path := lookupPath(FirefoxBinarySelenium3())
	firefoxBinarySelenium3 = &path
}
