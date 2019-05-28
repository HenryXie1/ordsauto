# Automation tool to create Http Ords and Loadbalancer in K8S

A kubectl plugin that create http and ords( Oracle Rest Data Services) based on [Apex (oracle application express) 19.1](https://github.com/HenryXie1/apexauto)


### Intro
Once we have [Apex](https://github.com/HenryXie1/apexauto) ready .  We often need to provision a http and ords for it. We would like to automate http ords and loadbalancer deployment in K8S. Once we have db hostname, port , sys password , apex /ords password. We can deployment a brand new http ords and loadbalancer deployment env via 1 command.  We can also delete it via 1 command.
ords image is based on docker images of [oracle github](https://github.com/oracle/docker-images).

### Demo
![Demo!](images/kubectl-apex-create1.gif)

## Installation

Download kubectl via [official guide](https://kubernetes.io/docs/tasks/tools/install-kubectl/) and configure access for your kubernetes cluster. Confirm kubectl get nodes is working

Download binary from [release link](https://github.com/HenryXie1/ordsauto/releases/download/v1.0/kubectl-ords)
Save it to /usr/local/bin of linux box (only linux supported as for now), No installation needed, download and run   
### Usage
```
$kubectl-ords
Usage:
  ords list|create|delete [-o ordsname][-n namespace][-d dbhostname] [-p 1521] [-s dbservice] [-w syspassword] [-x apexpassword]  [flags]

Examples:

  # Requirment: 
  # Install Oracle Apex in DB before running this tool. Refer automation tool for Apex
  #
  #
  # Docker images are based on Oracle GitHub Docker Repo 
  # ords container uses iad.ocir.io/espsnonprodint/autostg/apexords:v19
  # httpd container uses iad.ocir.io/espsnonprodint/autostg/oel-httpd:v4
  # list ords deployment with label app=peordshttp
  # list versions ords and Apex status in DB
        kubectl ords list -d dbhost -p 1521 -s testpdbsvc -w syspassword
        # create ords and http Pod with spefified name, run java ords.war install in the pod
        kubectl ords create -o myordsauto -d dbhost -p 1521 -s testpdbsvc -w syspassword -x apexpassword
        # delete ords deployment and drop ords related schemas in DB
        kubectl ords delete -o myordsauto -d dbhost -p 1521 -s testpdbsvc -w syspassword


Flags:
  -x, --apexpassword string   password for apex,ords related DB schemas
  -d, --dbhost string         DB hostname or IP address
  -p, --dbport string         DB port to access (default "1521")
  -h, --help                  help for ords
  -n, --namespace string      namespace for ords http deployment (default "default")
  -o, --ordsname string       name for ords http deployment
  -s, --service string        DB service to access
  -w, --syspassword string    sys password of DB service
```

### Contribution
More than welcome! please don't hesitate to open bugs, questions, pull requests 
