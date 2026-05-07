BINARY=terraform-provider-ldap
PLUGIN_PATH=$(HOME)/.terraform.d/plugins/registry.terraform.io/elastic-infra/ldap/$(shell git describe --tags | tr -d v)/darwin_amd64
TEST_ENV := LDAP_HOST=localhost LDAP_PORT=389 LDAP_BIND_USER="cn=admin,dc=example,dc=com" LDAP_BIND_PASSWORD=admin

.DEFAULT_GOAL: $(BINARY)

$(BINARY):
	go build -o bin/$(BINARY)

install: $(BINARY)
	install -d $(PLUGIN_PATH)
	install -m 775 bin/$(BINARY) $(PLUGIN_PATH)/

test:
	go test -v ./...

testacc:
	TF_ACC=1 go test -v ./...

docker_test:
	$(TEST_ENV) go test -v ./...

docs:
	@set -e; \
	go build -o /tmp/$(BINARY) .; \
	TF_DIR=$$(mktemp -d); \
	SCHEMA_JSON=$$(mktemp --suffix=.json); \
	printf 'terraform {\n  required_providers {\n    ldap = { source = "elastic-infra/ldap" }\n  }\n}\n' > $$TF_DIR/main.tf; \
	printf 'provider_installation {\n  dev_overrides { "elastic-infra/ldap" = "/tmp" }\n  direct {}\n}\n' > $$TF_DIR/terraform.rc; \
	TF_CLI_CONFIG_FILE=$$TF_DIR/terraform.rc terraform -chdir=$$TF_DIR providers schema -json \
		| python3 -c "import sys,json; d=json.load(sys.stdin); s=d['provider_schemas']; k=next(iter(s)); s['ldap']=s.pop(k); print(json.dumps(d))" \
		> $$SCHEMA_JSON; \
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest generate \
		--providers-schema $$SCHEMA_JSON; \
	rm -f /tmp/$(BINARY) $$SCHEMA_JSON; \
	rm -rf $$TF_DIR
