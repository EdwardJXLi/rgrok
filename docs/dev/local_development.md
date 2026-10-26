# Set up your development environment

rgrok runs as one binary and supports development on Linux, macOS, and Windows.
The included Nix shell provides the tools for the Task and Overmind workflow.
The default SQLite backend needs no separate database server.

## Step 1: Install dependencies

Use these tools:

- [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) v2 or higher.
- [Go](https://go.dev/doc/install) v1.26 or higher.
- [Node.js](https://nodejs.org/) v24 or higher.
- [pnpm](https://pnpm.io/installation) v11, as pinned in `package.json`.
- [Moon](https://moonrepo.dev/docs/install) v2 or higher, or [Task](https://taskfile.dev/installation/) v3 with [Overmind](https://github.com/DarthSim/overmind#installation) v2.
- PostgreSQL or MySQL, if you use either backend. The integration suite needs PostgreSQL.

On macOS, install the tools with Homebrew:

```bash
brew install git go node pnpm moon go-task overmind
```

For the Task and Overmind workflow, enter the Nix shell:

```bash
nix-shell
```

## Step 2: (Optional) Initialize a PostgreSQL database

If you use `database.type: postgres`, create the database and user.
If you use SQLite, skip this step.

```bash
createuser --superuser rgrokd
psql -c "ALTER USER rgrokd WITH PASSWORD 'rgrokd';"
createdb --owner=rgrokd --encoding=UTF8 --template=template0 rgrokd
```

## Step 3: Get the code

```bash
# HTTPS
git clone --depth 10 https://github.com/EdwardJXLi/rgrok.git

# or SSH
git clone --depth 10 git@github.com:EdwardJXLi/rgrok.git
```

> [!NOTE]
> Clone the repository outside your `$GOPATH`.

## Step 4: Initialize `rgrokd.yml`

Create `rgrokd.yml` in the repository root with this configuration:

```yaml
external_url: "http://localhost:3320"
web:
  port: 3320
proxy:
  port: 3000
  scheme: "http"
  domain: "localhost:3000"
sshd:
  port: 2222

database:
  type: "sqlite"
  path: "./rgrokd.db"

identity_provider:
  type: "oidc"
  display_name: "OIDC"
  issuer: "http://localhost:9833"
  client_id: "winnerwinner"
  client_secret: "chickendinner"
  field_mapping:
    identifier: "email"
    display_name: "name"
    email: "email"
```

For the PostgreSQL backend, replace the `database:` block with:

```yaml
database:
  type: "postgres"
  host: "localhost"
  port: 5432
  user: "rgrokd"
  password: "rgrokd"
  database: "rgrokd"
```

## Step 5: Start the servers

Install the web dependencies:

```bash
pnpm install --frozen-lockfile
```

Start the web, proxy, SSH, Vite, and mock OIDC servers:

```bash
moon run rgrokd:dev
```

Visit http://localhost:3320. The rgrokd web server forwards asset requests to Vite.
After you change Go source, run `moon run rgrokd:dev` again to rebuild the server.

For automatic Go rebuilds with Task and Overmind, use:

```bash
overmind start
```

To embed the web app in a binary with the `prod` build tag, use:

```bash
moon run rgrokd:prod
```

This command also starts the mock OIDC server.
