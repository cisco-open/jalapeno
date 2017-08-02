Attendees: Matt C, Adam S, Mike N, Omar D

## Overview:
- Keep it simple on first deploy and break out as needed.
- Initial design/use-case is on-prem and AWS is not a requirement for first pass...we can run on local openshift cluster.
  - Will also build an openshift on AWS for future testing but this is not a phase1 task.
- Stateless (no pet sets) for services..
- Stateful will be DB and Kafka (since if it dies you want to preserve those messages on restart...Kafka will die)


## Openshift pieces 
- Will leverage openshit routers pointing to service IPs since it provides all the HA/LB needs.  As we scale out we can explore if an nginx container layer will be needed to be less tied to openshift routers...pros/cons to this.
- Will leverage the hardened, and cisco approved, IVP-COE since it's a project with use-cases on it that align closely with voltron.
- We will need multiple openshift environments: dev, on-prem, off-prem (aws/cloud).
  -- The environments can be shared so multiple teams can have their own projects without stomping one another.




## Dev team
- Core dev: Mike Napolitano and Steve Louie (any west coast resources?)
  - Dev Consulting: Matt Caulfield
- Go language will be initial approach and neo4j/Kafka connectors exist
- A single container with various controllers for the various services and as needed we can break out into other containers:
  - In future we might need only a handful of api containers, but more DB containers (the neo4j-kafka connector).  Initially let's keep it simple as we find new issues.
- There will be some effort in learning neo4j and it's version of sql (cypherDB).





## Message Bus and DB components
- Initial Kafka will be a monolithic container. Kafka in containers (TBD) due to Zookeeper dependencies need to see how well this works...in practice it should but need to find some real world use-cases.  The big bottleneck with Kafka is fast disk IO so containers or not might not be the bottleneck.
- Initial neo4j will be a monolithic simple container and as we scale out we can explore if it will be a standalone VM or if neo4j on containers works well/stable 



## Cloud/off-prem infra
- I use off-prem here since it could be co-located at another customer DC and not in a Cloud provider so just keeping options open.
- Kafka pushing to two places: local and AWS (aws not first phase)
- Off-prem infra will be similar to on-prem but without collectors.
- Will cloud 


## Future Voltron
- What about deep analytics?  Are we thinking some kind of datalake?
- Spark streaming for any logging/event pre/post processing?
