# watch-dog
A Golang script that automaticaly post attendances for apprentices, based on school's access control  
  
## How to use  
1. `git clone ssh://git@gitlab.42nice.fr:4222/42nice/watch-dog.git`  
2. `cp config-default.yml ./config.yml` and fill it with your credentials  
3. `go run main.go ./config.yml \<path_to_log_folder\>`  
  
## What will happen  
Watch-dog will use the Access Control API to fetch every event of the current day. (Between 7:30 AM and 8:30 PM)  
It will then remove students that doesn't meet the following rules:  
- Didn't stayed more than 10 minute in school  
- Didn't had a correct 42Login and 42ID setted up in the Access Controle Database  
- Isn't subscribed to at least one alternant project on APIv2  
  
If AutoPost is set to `true` in config, it will then post an Attendance for every remaining students, on Chronos API.  
The Attendance will start with the first AC event, and stop with the last AC event of the given student.  
Source will be "access-control".  
  
Since AccessControle, v2 and Chronos API has read and write limits, this script can take several minutes to run.  
If you cancel the script before last step, no attendances will be posted.  

It will automaticly create a logfile in the folder you provide, using the date as filename.

## When to use
This script is meant to be run automaticly every day, at the end of the day. (After 20h30)  

## How to maintain
Since 42 API is not complete, we have no way to know which project are considered as a "Apprenticeship" project.  
In config file, you must provide a complete list of the projects IDs you want to track.   
The IDs provided in `config-default.yml` are the one that were active when the script got written.  
Double check the projects are the one used in your campus.  