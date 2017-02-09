bindir=bin
exe=echo_svr echo_clt pb_svr pb_clt

echo_svr_go=echo/server/*.go
echo_clt_go=echo/client/*.go

pb_svr_go=pb/server/*.go
pb_clt_go=pb/client/*.go

all:$(exe)

clean:
	rm -rf $(bindir)/*


echo_svr: $(echo_svr_go)
	@echo "building $@"
	go build -o $(bindir)/$@ $^

echo_clt: $(echo_clt_go)
	@echo "building $@"
	go build -o $(bindir)/$@ $^

pb_svr: $(pb_svr_go)
	@echo "building $@"
	go build -o $(bindir)/$@ $^

pb_clt: $(pb_clt_go)
	@echo "building $@"
	go build -o $(bindir)/$@ $^

