import yaml
import copy

with open('api-gateway/openapi.yaml', 'r') as f:
    spec = yaml.safe_load(f)

class NoAliasDumper(yaml.SafeDumper):
    def ignore_aliases(self, data):
        return True

with open('api-gateway/openapi.yaml', 'w') as f:
    yaml.dump(spec, f, sort_keys=False, allow_unicode=True, Dumper=NoAliasDumper)

