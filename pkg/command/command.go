package command

import (
	"github.com/patrickjahns/openvpn_exporter/pkg/collector"
	"github.com/patrickjahns/openvpn_exporter/pkg/version"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
)

// Run parses the command line arguments and executes the program.
func Run() error {

	app := &cli.App{
		Name:    "openvpn_exporter",
		Version: version.Info(),
		Usage:   "OpenVPN exporter",
		Authors: []*cli.Author{
			{
				Name:  "Patrick Jahns",
				Email: "github@patrickjahns.de",
			},
		},
	}

	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "Show help",
	}

	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Prints the current version",
	}

	app.Action = func(c *cli.Context) error {
		return run(c)
	}

	return app.Run(os.Args)
}

func run(c *cli.Context) error {
	// hardcoded vars for development, will be replaced with cli/config
	addr := ":9000"
	statusFile := "./example/version1.status"

	// setup logging
	logger := setupLogging()
	level.Info(logger).Log(
		"msg", "Starting openvpn_exporter",
		"version", version.Version,
		"revision", version.Revision,
		"buildDate", version.BuildDate,
		"goVersion", version.GoVersion,
	)

	// enable profiler
	r := prometheus.NewRegistry()
	r.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	r.MustRegister(prometheus.NewGoCollector())
	r.MustRegister(collector.NewGeneralCollector(
		logger,
		version.Version,
		version.Revision,
		version.BuildDate,
		version.GoVersion,
		version.Started,
	))
	r.MustRegister(collector.NewOpenVPNCollector(
		logger,
		"udp",
		statusFile,
	))
	http.Handle("/metrics",
		promhttp.HandlerFor(r, promhttp.HandlerOpts{}),
	)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>OpenVPN Exporter</title></head>
			<body>
			<h1>OpenVPN exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`))
	})

	level.Info(logger).Log("msg", "Listening on", "addr", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		level.Error(logger).Log("msg", "http listenandserve error", "err", err)
		return err
	}
	return nil
}

func setupLogging() log.Logger {
	filterOption := level.AllowDebug()
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewFilter(logger, filterOption)
	logger = log.With(logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)
	return logger
}
