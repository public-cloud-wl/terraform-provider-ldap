BINARY=terraform-provider-ldap
TEST_ENV := LDAP_HOST=localhost LDAP_PORT=389 LDAP_BIND_USER="cn=admin,dc=example,dc=com" LDAP_BIND_PASSWORD=admin
VERSION  := v1.0.2

.DEFAULT_GOAL: $(BINARY)

$(BINARY):
	go build -o bin/$(BINARY)

bootstrap:
	go install github.com/mitchellh/gox

cross-build:
	rm -rf dist
	CGO_ENABLED=0 GOFLAGS="-trimpath" gox -osarch "linux/amd64 darwin/amd64" \
		-output "dist/{{.OS}}_{{.Arch}}/$(BINARY)_$(VERSION)" \
		-ldflags "-w -s"
	tar cvzf $(BINARY)_$(VERSION).tar.gz -C dist .

test:
	go test -v

docker_test:
	$(TEST_ENV) go test -v
