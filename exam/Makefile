bindir=bin
exe=builder runner scheduler

builder_go=builder/*.go
runner_go=runner/*.go
scheduler_go=scheduler/*.go

all:$(exe)


builder: $(builder_go)
	@echo "building $@"
	go build -o $(bindir)/$@ $^

runner: $(runner_go)
	@echo "building $@"
	go build -o $(bindir)/$@ $^

scheduler: $(scheduler_go)
	@echo "building $@"
	go build -o $(bindir)/$@ $^

