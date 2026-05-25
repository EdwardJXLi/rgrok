# Deploy with single binary

1. Create a `rgrokd.yml` file:
    ```yaml
    external_url: "http://example.com"
    web:
      port: 3320
    proxy:
      port: 3000
      scheme: "http"
      domain: "example.com"
    sshd:
      port: 2222

    database:
      host: "localhost"
      port: 5432
      user: "REDACTED"
      password: "REDACTED"
      database: "rgrokd"

    identity_provider:
      type: "oidc"
      display_name: "Google"
      issuer: "https://accounts.google.com"
      client_id: "REDACTED"
      client_secret: "REDACTED"
      field_mapping:
        identifier: "email"
        display_name: "name"
        email: "email"
    #  # The required domain name, "field_mapping.email" is required to set for this to work.
    #  required_domain: "example.com"
    ```
1. Download the latest version of the `rgrokd` archive from the [Releases](https://github.com/EdwardJXLi/rgrok/releases) page.
1. Launch the `rgrokd` in background (systemd, screen, nohup).
    1. By default, `rgrokd` expects the `rgrokd.yml` is available in the working directory. Use `--config` flag to specify a different path for the config file.
