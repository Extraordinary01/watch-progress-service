# Watch progress service

## Prerequisites
Install Git, Go, make, Docker and Docker compose in your system


## Installation
1. Copy content of env.example file into .env file.
2. Run the command above:
```shell
go mod tidy
```

## Launch the server
```shell
make start-docker
make run
```
After stopping server don't forget to stop docker containers too.

### NOTES:
* For stopping server press Ctrl+C
* For development, you will need to set up your local ScyllaDB. To import service data you need to get schema.cql file 
from me (Begli Geldiyev).