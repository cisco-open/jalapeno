# Box for all non-markdown documents
[Box](https://cisco.box.com/s/5sgypvxyn08ushwzgqkrrien44eb6evb)
- [Use Cases](https://cisco.box.com/s/lrx2ch1lw6uea3ju8adxzf6ylljw9zb0)
- [CI Plan](https://cisco.box.com/s/9ux8j18zl1fg9p4t97kpij6hwzsqe9uk)
- [Architecture](https://cisco.box.com/s/vkjtjamvssxhnsejgesffrxcx3df0grj)
- [Goals](https://cisco.box.com/s/x9q40igydif1qsocbqqmkj4a5lruab0v)

# Repository details

### Docker Repository
We have created two Docker V2 Artifactory repositories for you. You can find the details below.  

Repo name : **gspie-docker** & **gspie-dev-docker**  

nsa-registry-dev-docker is a development snapshot repo which you can use to store your development artifacts. This dev repo will have regular clean up every 3 months as per our purge policy. Any unused artifact older than 3 months will be deleted and nsa-registry is a release repo where you will be storing all your releases.  

- Docker Registry URL: dockerhub.cisco.com/uccegen & dockerhub.cisco.com/gspie-dev-docker  
- Repo View URL: https://engci-maven.cisco.com/artifactory/list/gspie-docker  
- To Browse repository from Artifactory UI: http://engci-maven.cisco.com/artifactory/webapp/#/artifacts/browse/tree/General/gspie-docker  
- Index view: http://engci-maven.cisco.com/artifactory/list/gspie-docker  
- JFrog Documentation: http://www.jfrog.com/confluence/display/RTF/Docker+Repositories  
- EngCI Documentation: https://engit.cisco.com/ci/#/ci/dockerhub/docs  
- Open a case with [EngIT](https://youtu.be/dQw4w9WgXcQ)

Deployer Credentials: (use this account if you need to deploy artifacts through automation like Jenkins or a script)  

Username: gspie-deployer  
Password: 1mtj33p66ighsxs2  

#### Basic Usage  
1. Initial Setup  
Requires a Docker client running version 1.6 or newer  
Enter credentials for the deployer user:  
    docker login -u gspie-deployer -p [password here] -e  <cec>@cisco.com dockerhub.cisco.com  
2. Deploy  
docker pull hello-world  
docker tag hello-world dockerhub.cisco.com/gspie-dev-docker/hello-world  
docker tag hello-world dockerhub.cisco.com/gspie-docker/hello-world  
docker push dockerhub.cisco.com/gspie-dev-docker/hello-world  
docker push dockerhub.cisco.com/gspie-docker/hello-world  
3.  Download  
docker pull dockerhub.cisco.com/gspie-dev-docker/hello-world  
docker pull dockerhub.cisco.com/gspie-docker/hello-world  

### Generic Binary Repository

Here are the three repositories, use these URLs for uploads (_for generic binaries like XRv images_)

- Release Repo: http://engci-maven.cisco.com/artifactory/generic-release
- Snapshot Repo: http://engci-maven.cisco.com/artifactory/generic-snapshot
- Thirdparty Repo: http://engci-maven.cisco.com/artifactory/generic-thirdparty

The Release repo stores release artifacts, the Snapshot repo stores development artifacts, and the Thirdparty repo is used for storing any outside build dependencies your software needs.

**Use this URL for downloads**

Virtual Repo (above three repos combined): http://engci-maven.cisco.com/artifactory/generic-group

If you have any questions about using the application please do not hesitate to contact us via engit-ci-support@cisco.com – I would also be glad to meet with you over Webex if needed to discuss any details.   
