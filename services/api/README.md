# API Gateway
The API Gateway decouples direct backend interaction from anything external to Voltron's immediate data processing/topology generation. Using Voltron will be fielded by this API Gateway.

The API Gateway is documented via the Open API specification. Swagger UI is deployed alongside the API itself for easy documentation and testing.

The API Gateway and UI will be visible via NodePort `30880` and `30881`, per configuration in `*_np.yaml`.

## Usage
### Generate Swagger Code
In ```/voltron/services/api:```
1. Edit ```swagger.yaml```. This is the place to create new API endpoints or edit the parameters of previously created endpoints.
2. Execute ```./run_editor.sh```
3. Head to ```http://localhost:8080```. Import the ```swagger.yaml``` file you edited in step 1.
4. If you have no erorrs, generate the python-flask swagger server. This will download a zip file.
5. Copy the generated python-flask zip file into your ```/voltron/services/api/``` directory. Name it server.zip. Something like this: ```cp ~/Downloads/python-flask-server-generated.zip server.zip```
6. Execute ```./swagger_zip_to_src.sh server.zip```
### Edit generated python files according to old python originals
1. ```cd /api/src/swagger_server/controllers```
2. Copy function calls from old ```topology_controller``` and ```pathing_controller``` python files into new python files.
3. Add new calls as necessary. 
3. Head to ```/api/src/voltron/controllers/```
4. Add parameter validation in files there and make call to queries backend.
5. Head to ```/api/src/voltron/```.
6. Edit ```queries.py``` to add ArangoDB query logic.

Finally, build the image!


## Development
Development of the API is assisted by Swagger Editor via Docker, and custom code on the backend is hand-written.

### Swagger
Swagger is a useful API development tool which allows specification of an API in YAML corresponding to the Open API Specification. Swagger Editor is a web frontend for this specification, and also contains codegen abilities via Swagger Codegen to automatically generate an API structure, which lacks logic that must be filled in.

`run_editor.sh` will deploy Swagger Editor via Docker for development. http://localhost:8080/ is its deployed URL. The Swagger specification can be loaded in and then edited, and then saved back. `swagger.yml` in this folder is intended to be the master copy.

### Implementation
The code which contains real implementation is contained in `src/voltron`. `src/swagger_server/controllers/*.py` methods must then be linked to these implementations.

When a new Swagger server is generated, place the `.zip` into this folder, and run `./swagger_zip_to_src.sh` which will unzip and transplant the newest generated code into the `src/` folder. Original files in `src/swagger_server/controllers/` will be renamed `.py -> .old.py`. You will have to merge the `.old.py` code to the new `.py` functionality, and delete `.old.py`. Ideally this should be relatively simple as most implementation code should be decoupled into the `src/voltron/` folder.
