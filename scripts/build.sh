# #!/bin/bash

# command="docker-compose -f deployments/docker-compose.yml " # service
# command+="-f deployments/postgresql.yml "  # database
# command+="up --build"

# # run command
# $command

# ## Start
# # docker-compose -p microservice1 -f deployments/docker-compose.yml -f deployments/postgresql.yml up --build -d
# ## Stop
# # docker-compose -p microservice1 -f deployments/docker-compose.yml -f deployments/postgresql.yml stop
# ## Remove
# # docker-compose -p microservice1 -f deployments/docker-compose.yml -f deployments/postgresql.yml down