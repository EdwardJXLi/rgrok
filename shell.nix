{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  packages = with pkgs; [
    go
    gopls
    gotools
    go-tools
    golangci-lint
    delve

    nodejs_22
    pnpm

    go-task
    overmind

    postgresql_15

    git
  ];

  shellHook = ''
    export PGDATA="$PWD/.pg"
    export PGHOST="$PGDATA"
    export PGUSER="pgrokd"
    export PGDATABASE="pgrokd"

    echo -e "\033[38;5;2mrgrok environment active!\033[0m"
    echo "  go         $(go version | awk '{print $3}')"
    echo "  pnpm       $(pnpm --version)"
    echo "  task       $(task --version)"
    echo "  overmind   $(overmind --version | head -1)"
    echo "  postgres   $(postgres --version | awk '{print $3}')"
    echo ""
    echo "First-time Postgres setup (local, no system service needed):"
    echo "  initdb -U pgrokd --auth=trust --encoding=UTF8"
    echo "  pg_ctl -l \"\$PGDATA/log\" -o \"--unix_socket_directories='\$PGDATA'\" start"
    echo "  createdb -h \"\$PGDATA\" -U pgrokd pgrokd"
    echo ""
    echo "Then: overmind start  (reads Procfile, visit http://localhost:3320)"

    export P10K_CUSTOM_CONTEXT="rgrok"
    export P10K_CUSTOM_COLOR="2"
    export P10K_CUSTOM_ICON=""
  '';
}
