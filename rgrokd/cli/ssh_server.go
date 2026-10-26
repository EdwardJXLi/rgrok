package main

import (
	"context"

	"unknwon.dev/x/logx"

	"github.com/EdwardJXLi/rgrok/internal/conf"
	"github.com/EdwardJXLi/rgrok/internal/database"
	"github.com/EdwardJXLi/rgrok/internal/reverseproxy"
	"github.com/EdwardJXLi/rgrok/internal/sshd"
)

func runSSHServer(
	ctx context.Context,
	logger *logx.Logger,
	sshdPort int,
	proxy conf.Proxy,
	db *database.DB,
	proxies *reverseproxy.Cluster,
) error {
	return sshd.Start(
		ctx,
		logger,
		sshdPort,
		proxy,
		db,
		proxies,
	)
}
