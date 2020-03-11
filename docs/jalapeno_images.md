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
Now that you've built and pushed an image up into dockerhub, log onto the CentosKVM.
```bash
# Assuming you're working on BruceDev
ssh [USERNAME]@10.200.99.7
[PASSWORD]
ssh centos@10.0.250.2
password: cisco

# If you built and pushed your image upto the dockerhub repo from your local machine, be sure to pull it into the Jalapeno environment.
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

If the service IS already running with a prior image, to test out your new image:
1. Head to the OpenShift UI: https://10.200.99.7:8443 (assumes you are working off of BruceDev)|| (credentials are admin/admin)
2. Click on "Applications" on the left sidebar, and then either "Deployments" or "StatefulSets" depending on which Jalapeno service you're trying to update.
If you're not sure which one, just check both -- it will be in one or the other.
3. Click on the name of your service -- for example, "API".
4. Click on the top right hand dropdown "Actions".
5. Click "Edit YAML"
6. Update the YAML file with the new image tag and SHA256 key. 
7. Click Save

You should be able to head over to Applications/Pods in the OpenShift UI to see the Jalapeno service coming up with the new image.

#### And just like that, you've successfully built, pushed, and used your own Jalapeno image!
