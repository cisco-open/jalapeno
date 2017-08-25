# Repository details

> We have created two Docker V2 Artifactory repositories for you. You can find the details below.  
>   
> Repo name : gspie-docker & gspie-dev-docker  
>   
> nsa-registry-dev-docker is a development snapshot repo which you can use to store your development artifacts. This dev repo will have regular clean up every 3 months as per our purge policy. Any unused artifact older than 3 months will be deleted and nsa-registry is a release repo where you will be storing all your releases.  
>   
> Docker Registry URL: dockerhub.cisco.com/uccegen & dockerhub.cisco.com/gspie-dev-docker  
> Repo View URL: https://engci-maven.cisco.com/artifactory/list/gspie-docker  
> To Browse repository from Artifactory UI: http://engci-maven.cisco.com/artifactory/webapp/#/artifacts/browse/tree/General/gspie-docker  
> Index view: http://engci-maven.cisco.com/artifactory/list/gspie-docker  
> JFrog Documentation: http://www.jfrog.com/confluence/display/RTF/Docker+Repositories  
> EngCI Documentation: https://engit.cisco.com/ci/#/ci/dockerhub/docs  
>   
> Deployer Credentials: (use this account if you need to deploy artifacts through automation like Jenkins or a script)  
>   
> Username: gspie-deployer  
> Password: [emailed separately]  
>   
> Basic Usage:  
> 1). Initial Setup  
> Requires a Docker client running version 1.6 or newer  
> Enter credentials for the deployer user:  
>     docker login -u gspie-deployer -p [password here] -e  <cec>@cisco.com dockerhub.cisco.com  
> 2). Deploy  
> docker pull hello-world  
> docker tag hello-world dockerhub.cisco.com/gspie-dev-docker/hello-world  
> docker tag hello-world dockerhub.cisco.com/gspie-docker/hello-world  
> docker push dockerhub.cisco.com/gspie-dev-docker/hello-world  
> docker push dockerhub.cisco.com/gspie-docker/hello-world  
> 3). Download  
> docker pull dockerhub.cisco.com/gspie-dev-docker/hello-world  
> docker pull dockerhub.cisco.com/gspie-docker/hello-world  
>     
> If you have any questions about using the application please do not hesitate to contact us via engit-ci-support@cisco.com – I would also be glad to meet with you over Webex if needed to discuss any details.   
