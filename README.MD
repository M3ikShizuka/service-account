##
### Start
```bash
docker-compose -p microservice-account-service -f deployments/standalone.yml -f deployments/service.yml -f deployments/postgresql.yml up --build -d
```
### Stop
```bash
docker-compose -p microservice-account-service -f deployments/standalone.yml -f deployments/service.yml -f deployments/postgresql.yml stop
```
### Remove
```bash
docker-compose -p microservice-account-service -f deployments/standalone.yml -f deployments/service.yml -f deployments/postgresql.yml down
```

## Docker commands
Docker Compose would read the file docker-compose.yml, parse it into a data structure in memory, validate where possible, and print back the reconstruction of that configuration file from its internal representation. If this is not possible due to errors, it would print an error message instead.
```bash
docker-compose config
```
Build images.  
Note: no rebuild is triggered on the update to the source code unless the --build flag is provided. 
```bash
docker-compose up --build
```

## PostgreSQL commands
Login to db postgres
```bash
su postgres
psql -h <host> -p <pass> -U <user> -d <db name>
```

Display ports settings:
```SQL
SELECT * FROM pg_settings WHERE name = 'port';
```