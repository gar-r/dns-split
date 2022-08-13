package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"git.okki.hu/garric/dns-split/config"
	"git.okki.hu/garric/dns-split/proxy"
	"git.okki.hu/garric/dns-split/router"
)

func main() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	args := config.DeclaredArgs()
	cfg, err := config.Parse(args)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse config")
	}

	if args.Verbose {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	server := &proxy.Server{
		Config: cfg,
		Router: &router.Router{
			Config: cfg,
		},
	}

	log.Info().Str("path", args.ConfigLocation).Msg("config loaded")
	log.Info().Str("addr", cfg.Addr).Msg("starting server")
	log.Fatal().Err(server.ListenAndServe()).Msg("fatal server error")
}
