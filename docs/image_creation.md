# Jalapeno Images

## Building an image
To build an image, locate the ```Makefile``` in the directory of whichever Jalapeno service you're working on.
This ```Makefile``` defines the image ID and tag number of the Jalapeno service being deployed. For example
```bash
REPO=iejalapeno
IMAGE=api
TAG=0.0.1.2
```

After you're finished developing your service code, to build its image open the ```Makefile``` and change the tag number to avoid overwriting pre-existing functional code.
Then, execute in your terminal: 
```bash
make container
``` 
If your code is error-free, the image will built.

## Pushing an image
```bash
# Log into docker (assuming you have permissions to the dockerhub repository)
docker login

# See the images you've built locally
docker images

# Locate the image you want to push and craft the following command
docker push [REPOSTIORY NAME]:[TAG]
docker push iejalapeno/api:0.0.3 # sample command
```

## Using an image

# If you built and pushed a new image upto the dockerhub repo from your local machine, be sure to pull it into the Jalapeno environment.
docker pull [REPOSTIORY NAME]:[TAG]
docker pull iejalapeno/api:0.0.3 # sample command
```
To use the image, find the core YAML file that defines how the Jalapeno service is deployed.
For example, in the API service, the image is loaded in ```api.yaml```.
In this YAML file, there should be a pre-existing image listed in the code.
If you don't see something similar to the following, you might be looking at the wrong YAML file.
```bash
containers:
      - name: api
        image: iejalapeno/api:0.0.1.2@sha256:e35d9ad6a3a10ad4d39c3310c29b460afbf43ed9efeaf1fd5041881dafb24357
```
In this YAML file, update the image tag and SHA256 key (which should be returned any time you pull or push the image).

If the service IS NOT already runnning with a prior image, run the following to deploy the Jalapeno service with the new image:
```bash
oc apply -f ./api.yaml # assumes you have the OpenShift CLI tool "oc" installed, this is currently installed on the CentosKVM on all Jalapeno servers
```

If you've merely updated an existing image and wish to deploy it in an existing Jalapeno cluster:
1. Go to your microk8s UI and navigate to deployments, click on the image you wish to update, and then click on the edit (pencil) icon
2. Update the YAML file with the new image tag and SHA256 key. 
3. Microk8s will automatically delete the old Pod and bring up a new one using the updated image.

#### And just like that, you've successfully built, pushed, and used your own Jalapeno image!
