package main

import (
	"github.com/charmbracelet/log"

	"github.com/EdwardJXLi/rgrok/internal/conf"
	"github.com/EdwardJXLi/rgrok/internal/database"
	"github.com/EdwardJXLi/rgrok/internal/reverseproxy"
	"github.com/EdwardJXLi/rgrok/internal/sshd"
)

func startSSHServer(logger *log.Logger, sshdPort int, proxy conf.Proxy, db *database.DB, proxies *reverseproxy.Cluster) {
	logger = logger.WithPrefix("sshd")
	err := sshd.Start(
		logger,
		sshdPort,
		proxy,
		db,
		proxies,
	)
	if err != nil {
		logger.Fatal("Failed to start server", "error", err)
	}
}
