# 📡 42 Watchdog Webhook

A Golang-based HTTP server that listens to your school's access control system and posts daily attendance data for apprentices.

---

## 🧱 Architecture

This project contains **two executables**:

- **watchdog-server**: The HTTP service that listens to access events and tracks apprentice activity.
- **watchdog-client**: A CLI tool used to interact with the server (start/stop listening, get status, etc.).

---

## 🚀 How to Use

### 1. Go in the project
```bash
cd live-attendance
```
### 2. Configure

```bash
cp config-default.yml config.yml
```

Fill config.yml with your 42 API credentials, and the list of apprenticeship project IDs for your campus.

### 3. Install the Server as a systemd Service

```bash
make server-install
```

This will:
- Build the watchdog-server binary
- Copy it to /usr/local/bin/
- Install your config to /etc/watchdog/config.json
- Create a systemd service (without starting it)

### 4. Start the Service

```bash
make server-start
```

### 5. (Optional) Follow the Logs

```bash
make server-logs
```

---

## 💡 What It Does

The watchdog-server listens to webhook events such as door accesses.  
For each user detected:

1. If it's the first time, it queries the 42 API.
2. It checks if the user is an **apprentice**, based on internship project subscription with ongoing status.
3. The event is logged.
4. When you stop the listener via watchdog-client, attendance can be posted automatically.

---
## 🖥️ Commands via watchdog-client

Interact with the server using the following commands:

```bash
./watchdog-client start
```
Enables the server to listen to incoming access control events.  
If the server is not in "start" mode, any incoming access hooks will be ignored.

```bash
./watchdog-client stop
```
Stops listening for events.  
You can also pass the `--post-attendance` flag:

- This will iterate through all users and reset their timers.
- If a user meets the criteria (is an apprentice), their attendance will be posted.

```bash
./watchdog-client status
```
Displays a summary of all registered users, showing their first and last access times to the school.
  

(Coming soon)
```bash
./watchdog-client notifiy
```
Sends a notification (e.g., via Discord or Slack webhook) to every apprentice who meets the criteria:  
- Must be an apprentice  
- Has accumulated less than 7 hours of school presence that day

You can also override the server URL with:

```bash
watchdog-client --url http://localhost:8080/commands \<command\>
```
---

## 🔧 Maintenance

The 42 API does **not** expose which projects are "apprenticeship" related.  
You must **manually maintain** the list of project IDs in your config.yml.

- The provided list in config-default.yml was valid at time of writing.
- Always double-check that your campus uses those same projects.

---

## 🧹 Server Management

Makefile commands for managing the service:

| Action               | Command                  |
|----------------------|--------------------------|
| Install the server   | make server-install      |
| Start the server     | make server-start        |
| Stop the server      | make server-stop         |
| Restart it           | make server-restart      |
| See its logs         | make server-logs         |
| Check status         | make server-status       |
| Reload systemd       | make server-reload       |
| Uninstall the server | make server-delete       |

Logs are stored in: /var/log/watchdog.log (You can configure this in the makefile)

---

## 📦 Dependencies

- Go 1.21+
- systemd (Linux)
- Access Control API token
- 42 API token
- 42 Chronos API token

---

## 📜 License

MIT — Made at 42 Nice by [@TheKrainBow](https://github.com/TheKrainBow)
