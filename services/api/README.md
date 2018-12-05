# API Gateway
The API Gateway decouples direct backend interaction from anything external to Voltron's immediate data processing/topology generation. Using Voltron will be fielded by this API Gateway.

The API Gateway is documented via the Open API specification. Swagger UI is deployed alongside the API itself for easy documentation and testing.

The API Gateway and UI will be visible via NodePort `30880` and `30881`, per configuration in `*_np.yaml`.

## Usage

```bash
# Develop Swagger specification
./run_editor.sh
# CodeGen python-flask, example cp to here below
cp ~/Downloads/python-flask-server-generated.zip server.zip
./swagger_zip_to_src.sh server.zip
# Edit all .py files according to .old.py originals
pushd src/swagger_server/
...
popd
# Go wild developing
pushd src/voltron
...
popd
# Containerize
make container
# Push container to ievoltron
```

## Development
Development of the API is assisted by Swagger Editor via Docker, and custom code on the backend is hand-written.

### Swagger
Swagger is a useful API development tool which allows specification of an API in YAML corresponding to the Open API Specification. Swagger Editor is a web frontend for this specification, and also contains codegen abilities via Swagger Codegen to automatically generate an API structure, which lacks logic that must be filled in.

`run_editor.sh` will deploy Swagger Editor via Docker for development. http://localhost:8080/ is its deployed URL. The Swagger specification can be loaded in and then edited, and then saved back. `swagger.yml` in this folder is intended to be the master copy.

### Implementation
The code which contains real implementation is contained in `src/voltron`. `src/swagger_server/controllers/*.py` methods must then be linked to these implementations.

When a new Swagger server is generated, place the `.zip` into this folder, and run `./swagger_zip_to_src.sh` which will unzip and transplant the newest generated code into the `src/` folder. Original files in `src/swagger_server/controllers/` will be renamed `.py -> .old.py`. You will have to merge the `.old.py` code to the new `.py` functionality, and delete `.old.py`. Ideally this should be relatively simple as most implementation code should be decoupled into the `src/voltron/` folder.
