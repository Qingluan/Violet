package Funcs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

// Service controls a locally-running Selenium subprocess.
type Service struct {
	port            int
	addr            string
	cmd             *exec.Cmd
	shutdownURLPath string

	display, xauthPath string
	xvfb               *selenium.FrameBuffer

	geckoDriverPath, javaPath string
	chromeDriverPath          string
	htmlUnitPath              string

	output io.Writer
}

type ServiceOption func(*Service) error

func (self *BaseBrowser) PhantomJSService(port int, proxy ...string) (s *Service, err error) {
	L("Start phantomjs:", port)
	opt := ""
	if proxy != nil {
		if strings.HasPrefix(proxy[0], "socks5://") {
			opt += fmt.Sprintf(" --proxy=%s", strings.TrimLeft(proxy[0], "socks5://"))
			opt += fmt.Sprintf(" --proxy-type=socks5")
		} else if strings.HasPrefix(proxy[0], "http") {
			fs := strings.SplitN(proxy[0], "://", 2)
			opt += fmt.Sprintf(" --proxy=%s", fs[1])
			opt += fmt.Sprintf(" --proxy-type=%s", fs[0])
		}
	}
	args := fmt.Sprintf(" --webdriver-logfile=log.txt --webdriver=127.0.0.1:%d%s", port, opt)
	L("PhantomJs:", args)
	cmd := exec.Command(self.Path, strings.Fields(args)...)
	cmd.Env = os.Environ()
	if self.loger != nil {
		cmd.Stdout = self.loger
		cmd.Stderr = self.loger
	}
	s, err = newService(cmd, "", port)
	if err != nil {
		return nil, err
	}
	if err := s.start(port); err != nil {
		L("start err:", err)
		return nil, err
	}
	return s, nil
}

func newService(cmd *exec.Cmd, urlPrefix string, port int, opts ...ServiceOption) (*Service, error) {
	s := &Service{
		port: port,
		addr: fmt.Sprintf("http://localhost:%d%s", port, urlPrefix),
		// output: outer,
	}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	cmd.Stderr = s.output
	cmd.Stdout = s.output
	cmd.Env = os.Environ()
	// TODO(minusnine): Pdeathsig is only supported on Linux. Somehow, make sure
	// process cleanup happens as gracefully as possible.
	if s.display != "" {
		cmd.Env = append(cmd.Env, "DISPLAY=:"+s.display)
	}
	if s.xauthPath != "" {
		cmd.Env = append(cmd.Env, "XAUTHORITY="+s.xauthPath)
	}
	s.cmd = cmd
	return s, nil
}

func (s *Service) start(port int) error {
	if err := s.cmd.Start(); err != nil {
		return err
	}

	for i := 0; i < 30; i++ {
		time.Sleep(time.Second)
		resp, err := http.Get(s.addr + "/status")
		if err == nil {
			resp.Body.Close()
			switch resp.StatusCode {
			// Selenium <3 returned Forbidden and BadRequest. ChromeDriver and
			// Selenium 3 return OK.
			case http.StatusForbidden, http.StatusBadRequest, http.StatusOK:
				return nil
			}
		}
	}
	return fmt.Errorf("server did not respond on port %d", port)
}

// Stop shuts down the WebDriver service, and the X virtual frame buffer
// if one was started.
func (s *Service) Stop() error {
	// Selenium 3 stopped supporting the shutdown URL by default.
	// https://github.com/SeleniumHQ/selenium/issues/2852
	if s.shutdownURLPath == "" {
		if err := s.cmd.Process.Kill(); err != nil {
			return err
		}
	} else {
		resp, err := http.Get(s.addr + s.shutdownURLPath)
		if err != nil {
			return err
		}
		resp.Body.Close()
	}
	if err := s.cmd.Wait(); err != nil && err.Error() != "signal: killed" {
		return err
	}
	if s.xvfb != nil {
		return s.xvfb.Stop()
	}
	return nil
}
