# Teras Teknologi

HN Clone but it's event driven powered microservices.

## Setup
```
mv .env.example .env # change the values accordingly
make setup
```

## Running
```
make run-<service_name>
```

Or just wanted to build it:
```
make build-<service_name>
```

Migration:
```
make migrate-<service_name>
```

Where service_name:
- account
- auth
- feeds

## Docs
Currently only available API docs for feeds service.

## Notes
Some services may does not shipped with independent worker that you can scale easily.
But for now, it's enough to run it with self-contained worker.

## And 
Makefile is your best friend!.
