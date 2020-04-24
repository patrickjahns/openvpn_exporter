package command

import (
	"github.com/patrickjahns/openvpn_exporter/pkg/collector"
	"github.com/patrickjahns/openvpn_exporter/pkg/version"
	"net/http"
	"os"

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
	// enable profiler
	r := prometheus.NewRegistry()
	r.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	r.MustRegister(prometheus.NewGoCollector())
	r.MustRegister(collector.NewGeneralCollector(
		version.Version,
		version.Revision,
		version.BuildDate,
		version.GoVersion,
		version.Started,
	))
	r.MustRegister(collector.NewOpenVPNCollector("./example/version1.status"))
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
	if err := http.ListenAndServe(":9000", nil); err != nil {
		return err
	}
	return nil
}
