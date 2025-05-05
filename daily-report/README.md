# 📅 42 Watchdog - Daily Attendance Script

A standalone Golang script that automatically posts daily attendances for apprentices, based on access control data from your school.

---

## 🧱 What It Does

This tool fetches all access control events for the current day (between 07:30 and 20:30), filters them based on predefined rules, and optionally posts validated attendances to the Chronos API.

Steps performed:

1. Fetch access control events
2. Filter out users who:
   - Stayed less than 10 minutes
   - Don't have a valid 42 login or ID in the AC database
   - Are not subscribed to at least one active alternant project (checked via APIv2)
3. If `AutoPost` is enabled in your config:
   - Create and send attendances to Chronos
   - Attendance starts at the first access, ends at the last access
   - Source is set to "access-control"
4. A log file is generated in the folder you provide, named with the current date.

---

## 🚀 How to Use

### 1. Create and Edit Your Config

```bash
cp config-default.yml config.yml
```

Fill in your 42 API credentials, Chronos token, and the list of alternant project IDs for your campus.

### 2. Run the Script

```bash
go run main.go ./config.yml /path/to/log/folder
```

This will create a dated log file in the provided folder.

---

## ⏱️ When to Run It

This script is meant to be executed if the HTTP server in webhook folder failed to do his job.  

---

## 🔧 Maintenance

Because the 42 API does not expose which projects are considered "apprenticeship", you must maintain this manually.

- Edit `config.yml` and list all project IDs that qualify.
- The default list in `config-default.yml` was accurate at the time of writing.
- Verify that the project IDs match your campus's current alternant projects.

---

## 📦 Dependencies

- Go 1.21+
- Access Control API token
- 42 API v2 access
- 42 Chronos API token

---

## 📜 License

MIT — Made at 42 Nice by [@TheKrainBow](https://github.com/TheKrainBow)
