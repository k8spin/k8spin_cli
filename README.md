# K8Spin CLI
Version: Version: 0.1.14

# Get dependencies
```
go get github.com/olekukonko/tablewriter
go get github.com/parnurzeal/gorequest
go get github.com/urfave/cli
```

# Build
```
mkdir bin
export GOBIN=/home/pau/workspace/k8spin_cli/bin
export PATH=$PATH:$GOBIN
go install
```

# Release
Based on https://gitlab.com/deviosec/containers/release