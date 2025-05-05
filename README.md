# 🕵️ 42 Watchdog

Watchdog is a set of tools designed to automate apprentice attendance tracking based on your school's access control system.  
This repository contains **two complementary implementations**, each with its own purpose and usage.

---

## 📡 `live-attendance/` — Live HTTP Server

A Go server that listens in real time to access control webhooks and logs apprentice presence throughout the day.

- Runs 24/7 as a `systemd` service
- Can be controlled via the `watchdog-client` CLI
- Supports commands: start, stop (with optional attendance post), status, and notify

📄 See [`live-attendance/README.md`](live-attendance/README.md) for setup and usage instructions

---

## 📅 `daily-attendance/` — End-of-Day Script

A standalone Go script used **only if the server fails** to post attendances for a specific day.  
You manually specify the date, and it will reprocess the access control events and attempt to post the missing attendances.

- Not meant for daily use
- Useful to backfill or fix missing data
- Fully automated once the date is provided

📄 See [`daily-attendance/README.md`](daily-attendance/README.md) for setup and usage instructions

---

## 🛠️ Maintenance Reminder

Both implementations require manual maintenance of the list of alternant project IDs in their config file.  
Make sure these IDs are up to date with your campus's 42 curriculum.

---

MIT — Made at 42 Nice by [@TheKrainBow](https://github.com/TheKrainBow)
