package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/karust/openserp/baidu"
	"github.com/karust/openserp/bing"
	"github.com/karust/openserp/brave"
	"github.com/karust/openserp/core"
	"github.com/karust/openserp/duckduckgo"
	"github.com/karust/openserp/google"
	"github.com/karust/openserp/sogou"
	"github.com/karust/openserp/yandex"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

// rawEngine implements SearchEngine interface for raw HTTP requests
type rawEngine struct {
	name string
}

func (r *rawEngine) Search(q core.Query) ([]core.SearchResult, error) {
	// Inject proxy settings from config
	q.ProxyURL = config.App.ProxyURL
	q.Insecure = config.App.Insecure

	switch r.name {
	case "google":
		return google.Search(q)
	case "yandex":
		return yandex.Search(q)
	case "baidu":
		return baidu.Search(q)
	default:
		return nil, fmt.Errorf("unsupported engine: %s", r.name)
	}
}

func (r *rawEngine) SearchImage(q core.Query) ([]core.SearchResult, error) {
	return nil, fmt.Errorf("image search is not supported in raw mode for %s", r.name)
}

func (r *rawEngine) Name() string {
	return r.name
}

func (r *rawEngine) IsInitialized() bool {
	return true
}

func (r *rawEngine) GetRateLimiter() *rate.Limiter {
	// Use default rate limiter for raw requests
	return rate.NewLimiter(rate.Every(time.Second), 5)
}

var serveCMD = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"listen"},
	Short:   "Start HTTP server, to provide search engine results via API",
	Args:    cobra.MatchAll(cobra.NoArgs),
	Run:     serve,
}

func serve(cmd *cobra.Command, args []string) {
	if config.App.IsRawRequests {
		logrus.Warn("Browserless results are very inconsistent or may not even work!")
		serv := core.NewServer(config.App.Host, config.App.Port,
			&rawEngine{name: "google"},
			&rawEngine{name: "yandex"},
			&rawEngine{name: "baidu"},
			&rawEngine{name: "brave"},
		)
		_serve(serv)
		return
	}

	opts := core.BrowserOpts{
		IsHeadless:          !config.App.IsBrowserHead, // Disable headless if browser head mode is set
		IsLeakless:          config.App.IsLeakless,
		Timeout:             time.Second * time.Duration(config.App.Timeout),
		LeavePageOpen:       config.App.IsLeaveHead,
		CaptchaSolverApiKey: config.Config2Capcha.ApiKey,
		ProxyURL:            config.App.ProxyURL,
		Insecure:            config.App.Insecure,
		UseStealth:          config.App.IsStealth,
	}

	if config.App.IsDebug {
		opts.IsHeadless = false
	}

	browser, err := core.NewBrowser(opts)
	if err != nil {
		logrus.Error(err)
		return
	}

	// Ensure browser is closed when function returns (covers early returns)
	defer func() {
		if err := browser.Close(); err != nil {
			logrus.Errorf("Browser close error: %v", err)
		} else {
			logrus.Info("Browser closed")
		}
	}()

	yand := yandex.New(*browser, config.YandexConfig)
	gogl := google.New(*browser, config.GoogleConfig)
	baidu := baidu.New(*browser, config.BaiduConfig)
	bing := bing.New(*browser, config.BingConfig)
	ddg := duckduckgo.New(*browser, config.DuckDuckGoConfig)
	brave := brave.New(*browser, config.BraveConfig)
	sogou := sogou.New(*browser, config.SogouConfig)

	serv := core.NewServer(config.App.Host, config.App.Port, gogl, yand, baidu, bing, ddg, brave, sogou)
	_serve(serv)
}

func init() {
	RootCmd.AddCommand(serveCMD)
}

func _serve(serv *core.Server) {
	// Start server and listen for shutdown signals
	errCh := make(chan error, 1)
	go func() { errCh <- serv.Listen() }()

	logrus.Infof("HTTP server started on %s:%d", config.App.Host, config.App.Port)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	select {
	case sig := <-sigCh:
		logrus.Infof("Received signal %s, shutting down server and browser...", sig)
		if err := serv.Shutdown(); err != nil {
			logrus.Errorf("Server shutdown error: %v", err)
		} else {
			logrus.Info("Server shutdown completed")
		}
		// Browser will be closed by deferred call
	case err := <-errCh:
		if err != nil {
			logrus.Errorf("Server exited with error: %v", err)
		} else {
			logrus.Info("Server stopped")
		}
		// Browser will be closed by deferred call
	}
}
