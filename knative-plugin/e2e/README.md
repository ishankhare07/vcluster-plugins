### How to execute the e2e tests locally

1. Start the vcluster via `devspace run dev` and running `go run -mod vendor main.go` in the terminal
2. Then start the e2e tests via `VCLUSTER_SUFFIX=vcluster-knative VCLUSTER_NAMESPACE=vcluster-knative VCLUSTER_NAME=vcluster-knative go test -v -ginkgo.v`
