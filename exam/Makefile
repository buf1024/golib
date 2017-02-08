bindir=bin
exe=echo_svr echo_clt

echo_svr_go=echo/server/*.go
echo_clt_go=echo/client/*.go

all:$(exe)


echo_svr: $(echo_svr_go)
	@echo "building $@"
	go build -o $(bindir)/$@ $^

echo_clt: $(echo_clt_go)
	@echo "building $@"
	go build -o $(bindir)/$@ $^

