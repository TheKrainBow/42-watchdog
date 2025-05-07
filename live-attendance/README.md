# 📡 42 Watchdog Webhook

A Golang-based HTTP server that listens to your school's access control system and posts daily attendance data for apprentices.

---

## 🛠️ Prerequisites

* Go 1.21+
* systemd (Linux)
* A 42 API token
* A 42 Chronos API token
* Access control webhook support (from your school)

---

## 📇 Architecture Overview

This project contains **two executables**:

* **watchdog-server**: The main HTTP service that listens to webhook events and manages users.
* **watchdog-client**: A CLI tool used to interact with the server (start/stop tracking, notify students, get status, etc.).

---

## 🚀 Setup Instructions

### 1. Clone the repo & enter the project directory

```bash
cd live-attendance
```

### 2. Configure your credentials

```bash
cp config-default.yml config.yml
```

Edit `config.yml` with:

* Your 42 API token
* Your campus apprenticeship project IDs
* Discord/Slack webhook (if applicable)

### 3. Build locally (no install)

```bash
make local-build
```

This will compile the binaries to the current folder.

### 4. Install the server as a systemd service

```bash
make server-install
```

This will:

* Copy `watchdog-server` to `/usr/local/bin/`
* Copy your config to `/etc/watchdog/config.yml`
* Create a systemd unit at `/etc/systemd/system/watchdog.service`

### 5. Start the service

```bash
make server-start
```

### 6. View logs

```bash
make server-logs
```

### 7. Install autocompletion (optional)

```bash
make client-install-cmd-completion-zsh    # for zsh
make client-install-cmd-completion-bash   # for bash
```

### 8. Set up automated cron commands

```bash
make cron-setup
```

This installs:

* `watchdog-client start` at 07:30
* `watchdog-client notify` at 19:30
* `watchdog-client stop --post-attendance` at 20:30

---

## 🔎 Interacting with the server via CLI

```bash
watchdog-client start                   # Begin listening
watchdog-client stop                    # Stop listening
watchdog-client stop --post-attendance  # Stop & send attendance
watchdog-client status                  # View current user access states
watchdog-client notify                  # Notify students with low time
```

All commands are sent by default to `http://localhost:8080/commands` — override with:

```bash
watchdog-client --url <custom-url> <command>
```

---

## 🛠️ Makefile Targets

### 📁 Local build & clean

| Command            | Description                            |
| ------------------ | -------------------------------------- |
| `make local-build` | Build both binaries in the current dir |
| `make local-clean` | Remove local binaries                  |

### 📂 System install

| Command             | Description                             |
| ------------------- | --------------------------------------- |
| `make system-build` | Copy built binaries to `/usr/local/bin` |
| `make system-clean` | Remove binaries from `/usr/local/bin`   |

### ⚖️ Server management

| Command                 | Description                             |
| ----------------------- | --------------------------------------- |
| `make server-install`   | Install service and config              |
| `make server-start`     | Start and enable the service            |
| `make server-stop`      | Stop and disable the service            |
| `make server-restart`   | Stop, rebuild, and restart the service  |
| `make server-status`    | View current service status             |
| `make server-logs`      | View live logs                          |
| `make server-uninstall` | Remove the service & config (keep logs) |

### ⏰ Cron setup

| Command            | Description                            |
| ------------------ | -------------------------------------- |
| `make cron-setup`  | Add 3 daily client commands to crontab |
| `make cron-remove` | Remove all cron jobs for the client    |

### 🔢 Autocompletion

| Command                                     | Description             |
| ------------------------------------------- | ----------------------- |
| `make client-install-cmd-completion-zsh`    | Install zsh completion  |
| `make client-install-cmd-completion-bash`   | Install bash completion |
| `make client-uninstall-cmd-completion-zsh`  | Remove zsh completion   |
| `make client-uninstall-cmd-completion-bash` | Remove bash completion  |

### 🚮 Full cleanup

| Command      | Description                                                              |
| ------------ | ------------------------------------------------------------------------ |
| `make purge` | Prompt to delete all binaries, service, cron, autocompletion (logs stay) |

Use `make help` to list all commands grouped by category.

---

## 🔎 Notes

* `watchdog-server` logs to `/var/log/watchdog.log`
* You can adjust logging, config paths, and more from the `Makefile`

---

MIT — Made at 42 Nice by [@TheKrainBow](https://github.com/TheKrainBow)
