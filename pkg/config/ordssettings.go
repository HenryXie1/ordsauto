package config

var (
	OrdsExample = `
	# ords container uses iad.ocir.io/espsnonprodint/livesqlsandbox/apexords:v19
	# http container uses iad.ocir.io/espsnonprodint/livesqlsandbox/livesqlstg-ohs:v3
	# list ords deployment with label app=peordshttp
	kubectl ords list 
	# create ords and http Pod . 
	kubectl ords create -d dbhost -p 1521 -s testpdbsvc -w syspassword -x apexpassword
	# delete ords deployment with label app=peordshttp
	kubectl ords delete 
	`
	OrdsHttpyml = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: livesqlstg-ordshttp-deployment-auto
  labels:
    app: peordshttp
    name: livesqlstg-service
spec:
  replicas: 1
  selector:
      matchLabels:
         name: livesqlstg-service
  template:
       metadata:
          labels:
             name: livesqlstg-service
       spec:
         containers:
           - name: ords
             image: iad.ocir.io/espsnonprodint/livesqlsandbox/apexords:v19
             imagePullPolicy: IfNotPresent
             ports:
                - containerPort: 8888
           - name: httpd
             image: iad.ocir.io/espsnonprodint/livesqlsandbox/livesqlstg-ohs:v3
             imagePullPolicy: IfNotPresent
             ports:
                - containerPort: 80
         imagePullSecrets:
            - name: iad-ocir-secret
`
)

