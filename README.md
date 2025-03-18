
# Install Docker Desktop 
https://docs.docker.com/desktop/install/mac-install/

Start the docker app after the installment.

## Install go@1.20
```sh
   brew install go@1.20
   echo 'export PATH="/opt/homebrew/opt/go@1.20/bin:$PATH"' >> ~/.zshrc
```

# Run Local Unit Tests
```bash
# start docker compose
docker compose up -d

# run test on webhook node
go test -v service/workflow_service/nodes_test/init_test.go  service/workflow_service/nodes_test/webhook_test.go

# run test on if node
go test -v service/workflow_service/nodes_test/init_test.go service/workflow_service/nodes_test/if_test.go

# run test on code node
go test -v service/workflow_service/nodes_test/init_test.go service/workflow_service/nodes_test/code_test.go

# Shut down the docker compose
docker compose down
```


# FYI
Common developing commands
```sh
   go mod tidy
   go mod vendor
   make build-go
   make clean-go
```

