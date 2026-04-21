# SSH Portfolio

A lightweight SSH portfolio service that displays your public SSH key information via a custom SSH banner.

## Features

- Serves SSH public key information as a custom banner
- Built with Go — simple and efficient
- Runs as a systemd service
- Accessed via domain: `ssh.zulfifazhar.dev`

## Architecture

```
Client (SSH)
    │
    ▼
nginx (port 22)
    │ (stream proxy)
    ▼
ssh-portfolio (port 2222)
    │
    ▼
Returns SSH public key info as banner
```

## Access

```bash
ssh ssh.zulfifazhar.dev
```

## License

MIT
