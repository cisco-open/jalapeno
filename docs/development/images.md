# Build & Deploy Images

This page contains the steps to build Jalapeno docker images & publish to a docker respository.

## Building an image

To build an image, locate the `Makefile` in the directory of whichever Jalapeno service you're working on. This `Makefile` defines the image ID and tag number of the Jalapeno service being deployed.

For example:

```bash
REPO=iejalapeno
IMAGE=api
TAG=0.0.1.2
```

After you have finished developing your service code, open the `Makefile` and change the tag number to avoid overwriting pre-existing functional code.

Then, execute in your terminal:

```bash
make container
```

If there are no errors, the image will build.

## Pushing an image

To push a Jalapeno docker image to a docker repository:

1. Log into docker (assuming you have permissions to the dockerhub repository)

    ```bash
    docker login
    ```

2. List images you've built locally:

    ```bash
    docker images
    ```

3. Locate the image you want to push and use the following command, replacing the repository name & tag as needed:

    ```bash
    docker push [REPOSTIORY NAME]:[TAG]
    docker push iejalapeno/api:0.0.3 # sample command
    ```

## Using an image

Once you have built and pushed a new image to the repository, you can pull it into the Jalapeno environment:

```bash
docker pull [REPOSTIORY NAME]:[TAG]
docker pull iejalapeno/api:0.0.3 # sample command
```

To use the image, find the core YAML file that defines how the Jalapeno service is deployed. For example, in the API service, the image is loaded in `api.yaml`.
In this YAML file, there should be a pre-existing image listed in the code. This should look similar to the output below:

```bash
containers:
      - name: api
        image: iejalapeno/api:0.0.1.2@sha256:e35d9ad6a3a10ad4d39c3310c29b460afbf43ed9efeaf1fd5041881dafb24357
```

In this YAML file, update the image tag and SHA256 key (which should be returned any time you pull or push the image).

If the service IS NOT already runnning with a prior image, run the following to deploy the Jalapeno service with the new image:

```bash
oc apply -f ./api.yaml #(1)!
```

1. Assumes you have the OpenShift CLI tool "oc" installed, this is currently installed on the CentosKVM on all Jalapeno servers

If you've updated an image and wish to deploy it in an existing Jalapeno cluster:

1. Go to your microk8s web UI and navigate to deployments.
2. Click on the image you wish to update, and then click on the edit (pencil) icon
3. Update the YAML file with the new image tag and SHA256 key.
4. Microk8s will automatically delete the old Pod and bring up a new one using the updated image.

**And just like that, you've successfully built, pushed, and used your own Jalapeno image! 🎉**
