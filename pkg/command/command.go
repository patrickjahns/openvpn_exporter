package command

import (
	"net/http"
	"os"
	"strings"

	"github.com/patrickjahns/openvpn_exporter/pkg/openvpn"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"

	"github.com/patrickjahns/openvpn_exporter/pkg/collector"
	"github.com/patrickjahns/openvpn_exporter/pkg/config"
	"github.com/patrickjahns/openvpn_exporter/pkg/version"
)

// Run parses the command line arguments and executes the program.
func Run() error {
	app, cfg := initApp()

	app.Action = func(c *cli.Context) error {
		server, logger := run(cfg)
		if err := server.ListenAndServe(); err != nil {
			level.Error(logger).Log("msg", "http listenandserve error", "err", err)
			return err
		}
		return nil
	}

	return app.Run(os.Args)
}

func initApp() (*cli.App, *config.Config) {
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
	cfg := config.Load()
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

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "web.address",
			Aliases:     []string{"web.listen-address"},
			Value:       "0.0.0.0:9176",
			Usage:       "Address to bind the metrics server",
			EnvVars:     []string{"OPENVPN_EXPORTER_WEB_ADDRESS"},
			Destination: &cfg.Server.Addr,
		},
		&cli.StringFlag{
			Name:        "web.path",
			Aliases:     []string{"web.telemetry-path"},
			Value:       "/metrics",
			Usage:       "Path to bind the metrics server",
			EnvVars:     []string{"OPENVPN_EXPORTER_WEB_PATH"},
			Destination: &cfg.Server.Path,
		},
		&cli.StringFlag{
			Name:        "web.root",
			Value:       "/",
			Usage:       "Root path to exporter endpoints",
			EnvVars:     []string{"OPENVPN_EXPORTER_WEB_ROOT"},
			Destination: &cfg.Server.Root,
		},
		&cli.StringSliceFlag{
			Name:     "status-file",
			Usage:    "The OpenVPN status file(s) to export (example test:./example/version1.status )",
			EnvVars:  []string{"OPENVPN_EXPORTER_STATUS_FILE"},
			Required: true,
		},
		&cli.BoolFlag{
			Name:    "disable-client-metrics",
			Usage:   "Disables per client (bytes_received, bytes_sent, connected_since) metrics",
			EnvVars: []string{"OPENVPN_EXPORTER_DISABLE_CLIENT_METRICS"},
		},
		&cli.BoolFlag{
			Name: "pseudonymize-client-metrics",
			Usage: "Pseudonymized per client (bytes_received, bytes_sent, connected_since) metrics by replacing " +
				"usernames with a random string",
			EnvVars: []string{"OPENVPN_EXPORTER_PSEUDONYMIZE_CLIENT_METRICS"},
		},
		&cli.IntFlag{
			Name:    "pseudonymize-client-metrics-length",
			Value:   8,
			Usage:   "Length of the client pseudonym string",
			EnvVars: []string{"OPENVPN_EXPORTER_PSEUDONYMIZE_CLIENT_METRICS_LENGTH"},
		},
		&cli.BoolFlag{
			Name:        "enable-golang-metrics",
			Value:       false,
			Usage:       "Enables golang and process metrics for the exporter) ",
			EnvVars:     []string{"OPENVPN_EXPORTER_ENABLE_GOLANG_METRICS"},
			Destination: &cfg.ExportGoMetrics,
		},
		&cli.StringFlag{
			Name:        "log.level",
			Value:       "info",
			Usage:       "Only log messages with given severity",
			EnvVars:     []string{"OPENVPN_EXPORTER_LOG_LEVEL"},
			Destination: &cfg.Logs.Level,
		},
	}

	app.Before = func(c *cli.Context) error {
		cfg.StatusCollector.StatusFile = c.StringSlice("status-file")
		cfg.StatusCollector.ExportClientMetrics = !c.Bool("disable-client-metrics")
		cfg.StatusCollector.PseudonymizeClientMetrics = c.Bool("pseudonymize-client-metrics")
		cfg.StatusCollector.PseudonymizeClientMetricsLength = c.Int("pseudonymize-client-metrics-length")
		return nil
	}
	return app, cfg
}

func run(cfg *config.Config) (*http.Server, log.Logger) {
	// setup logging
	logger := setupLogging(cfg)
	level.Info(logger).Log(
		"msg", "Starting openvpn_exporter",
		"version", version.Version,
		"revision", version.Revision,
		"buildDate", version.BuildDate,
		"goVersion", version.GoVersion,
	)
	var openVPServers []collector.OpenVPNServer
	r := prometheus.NewRegistry()
	if cfg.ExportGoMetrics {
		// enable profiler
		r.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
		r.MustRegister(prometheus.NewGoCollector())
	}
	r.MustRegister(collector.NewGeneralCollector(
		logger,
		version.Version,
		version.Revision,
		version.BuildDate,
		version.GoVersion,
		version.Started,
	))
	for _, statusFile := range cfg.StatusCollector.StatusFile {
		serverName, statusFile := parseStatusFileSlice(statusFile)
		level.Info(logger).Log(
			"msg", "registering collector for",
			"serverName", serverName,
			"statusFile", statusFile,
		)
		openVPServers = append(openVPServers, collector.OpenVPNServer{Name: serverName, StatusFile: statusFile, ParseError: 0})
	}

	var parserDecorators []openvpn.ParserDecorator
	if cfg.StatusCollector.PseudonymizeClientMetrics {
		parserDecorators = append(
			parserDecorators,
			openvpn.NewOpenVPNPseudonymizingDecorator(
				cfg.StatusCollector.PseudonymizeClientMetricsLength,
			),
		)
	}
	r.MustRegister(collector.NewOpenVPNCollector(
		logger,
		openVPServers,
		parserDecorators,
		cfg.StatusCollector.ExportClientMetrics,
	))

	mux := http.NewServeMux()
	mux.Handle(cfg.Server.Path,
		promhttp.HandlerFor(r, promhttp.HandlerOpts{}),
	)
	mux.HandleFunc(cfg.Server.Root, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>OpenVPN Exporter</title></head>
			<body>
			<h1>OpenVPN exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`))
	})

	level.Info(logger).Log("msg", "Listening on", "addr", cfg.Server.Addr)
	server := &http.Server{Addr: cfg.Server.Addr, Handler: mux}
	return server, logger
}

func parseStatusFileSlice(statusFile string) (string, string) {
	parts := strings.Split(statusFile, ":")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return parts[0], parts[0]
}

func setupLogging(cfg *config.Config) log.Logger {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	switch strings.ToLower(cfg.Logs.Level) {
	case "error":
		logger = level.NewFilter(logger, level.AllowError())
	case "warn":
		logger = level.NewFilter(logger, level.AllowWarn())
	case "info":
		logger = level.NewFilter(logger, level.AllowInfo())
	case "debug":
		logger = level.NewFilter(logger, level.AllowDebug())
	default:
		logger = level.NewFilter(logger, level.AllowInfo())
	}
	logger = log.With(logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)
	return logger
}
