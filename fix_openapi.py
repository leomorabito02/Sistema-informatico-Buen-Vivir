import yaml

def main():
    with open('api-gateway/openapi.yaml', 'r') as f:
        spec = yaml.safe_load(f)
        
    if 'x-google-endpoints' in spec:
        del spec['x-google-endpoints']
        
    for path, methods in spec.get('paths', {}).items():
        if 'options' not in methods:
            methods['options'] = {
                'summary': f'CORS preflight for {path}',
                'operationId': f'options_{path.replace("/", "_").replace("{", "").replace("}", "")}',
                'x-google-backend': {
                    'address': '${AUTH_SERVICE_URL}',
                    'path_translation': 'APPEND_PATH_TO_ADDRESS'
                },
                'responses': {
                    '200': {
                        'description': 'OK'
                    }
                }
            }
            
    with open('api-gateway/openapi.yaml', 'w') as f:
        yaml.dump(spec, f, sort_keys=False, allow_unicode=True)
        
if __name__ == '__main__':
    main()
