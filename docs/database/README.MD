# Database
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