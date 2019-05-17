package config

var (
OrdsExample = `
	# ords container uses iad.ocir.io/espsnonprodint/livesqlsandbox/apexords:v19
	# http container uses iad.ocir.io/espsnonprodint/livesqlsandbox/livesqlstg-ohs:v3
	# list ords deployment with label app=peordshttp
	kubectl ords list 
	# create ords and http Pod . 
	kubectl ords create -d dbhost -p 1521 -s testpdbsvc -w syspassword
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
Ordsconfigmapyml = `
apiVersion: v1
data:
  apex.xml: |
    <?xml version="1.0" encoding="UTF-8" standalone="no"?>
    <!DOCTYPE properties SYSTEM "http://java.sun.com/dtd/properties.dtd">
    <properties>
    <comment>Saved on Fri May 17 00:59:48 GMT 2019</comment>
    <entry key="db.hostname">livesql-sb-dbvm1.subnet1ad1iad.espsocicorpniad.oraclevcn.com</entry>
    <entry key="db.password">@05A2A5D548861FDC3863120E33437F2C07B7C1DB10F350B8FD</entry>
    <entry key="db.port">1521</entry>
    <entry key="db.servicename">apextest</entry>
    <entry key="db.username">APEX_PUBLIC_USER</entry>
    </properties>
  apex_al.xml: |
    <?xml version="1.0" encoding="UTF-8" standalone="no"?>
    <!DOCTYPE properties SYSTEM "http://java.sun.com/dtd/properties.dtd">
    <properties>
    <comment>Saved on Fri May 17 00:59:48 GMT 2019</comment>
    <entry key="db.hostname">livesql-sb-dbvm1.subnet1ad1iad.espsocicorpniad.oraclevcn.com</entry>
    <entry key="db.password">@054DC6AF482E826EF4995C064BEAB61BDD8257723CEF7A2CD2</entry>
    <entry key="db.port">1521</entry>
    <entry key="db.servicename">apextest</entry>
    <entry key="db.username">APEX_LISTENER</entry>
    </properties>
  apex_pu.xml: |
    <?xml version="1.0" encoding="UTF-8" standalone="no"?>
    <!DOCTYPE properties SYSTEM "http://java.sun.com/dtd/properties.dtd">
    <properties>
    <comment>Saved on Fri May 17 00:59:48 GMT 2019</comment>
    <entry key="db.hostname">livesql-sb-dbvm1.subnet1ad1iad.espsocicorpniad.oraclevcn.com</entry>
    <entry key="db.password">@055BB4A740D94A21735E5998FB95617EE10F1883C4830BC5FC</entry>
    <entry key="db.port">1521</entry>
    <entry key="db.servicename">apextest</entry>
    <entry key="db.username">ORDS_PUBLIC_USER</entry>
    </properties>
  apex_rt.xml: |
    <?xml version="1.0" encoding="UTF-8" standalone="no"?>
    <!DOCTYPE properties SYSTEM "http://java.sun.com/dtd/properties.dtd">
    <properties>
    <comment>Saved on Fri May 17 00:59:48 GMT 2019</comment>
    <entry key="db.hostname">livesql-sb-dbvm1.subnet1ad1iad.espsocicorpniad.oraclevcn.com</entry>
    <entry key="db.password">@0540635D18DFAE4F4FBCF44A4D457863913D924879EAE12D53</entry>
    <entry key="db.port">1521</entry>
    <entry key="db.servicename">apextest</entry>
    <entry key="db.username">APEX_REST_PUBLIC_USER</entry>
    </properties>
  defaults.xml: |
    <?xml version="1.0" encoding="UTF-8" standalone="no"?>
    <!DOCTYPE properties SYSTEM "http://java.sun.com/dtd/properties.dtd">
    <properties>
    <comment>Saved on Fri May 17 00:59:48 GMT 2019</comment>
    <entry key="cache.caching">false</entry>
    <entry key="cache.directory">/tmp/apex/cache</entry>
    <entry key="cache.duration">days</entry>
    <entry key="cache.expiration">7</entry>
    <entry key="cache.maxEntries">500</entry>
    <entry key="cache.monitorInterval">60</entry>
    <entry key="cache.procedureNameList"/>
    <entry key="cache.type">lru</entry>
    <entry key="debug.debugger">false</entry>
    <entry key="debug.printDebugToScreen">false</entry>
    <entry key="error.keepErrorMessages">true</entry>
    <entry key="error.maxEntries">50</entry>
    <entry key="jdbc.DriverType">thin</entry>
    <entry key="jdbc.InactivityTimeout">1800</entry>
    <entry key="jdbc.InitialLimit">3</entry>
    <entry key="jdbc.MaxConnectionReuseCount">1000</entry>
    <entry key="jdbc.MaxLimit">10</entry>
    <entry key="jdbc.MaxStatementsLimit">10</entry>
    <entry key="jdbc.MinLimit">1</entry>
    <entry key="jdbc.statementTimeout">900</entry>
    <entry key="log.logging">false</entry>
    <entry key="log.maxEntries">50</entry>
    <entry key="misc.compress"/>
    <entry key="misc.defaultPage">apex</entry>
    <entry key="security.disableDefaultExclusionList">false</entry>
    <entry key="security.maxEntries">2000</entry>
    <entry key="security.requestValidationFunction">wwv_flow_epg_include_modules.authorize</entry>
    <entry key="security.validationFunctionType">plsql</entry>
    </properties>
  ords_params.properties: |
    db.hostname=livesql-sb-dbvm1.subnet1ad1iad.espsocicorpniad.oraclevcn.com
    db.password=BFE2GRPF
    db.port=1521
    db.servicename=apextest
    db.username=APEX_PUBLIC_USER
    migrate.apex.rest=false
    plsql.gateway.add=true
    rest.services.apex.add=true
    rest.services.ords.add=true
    schema.tablespace.default=SYSAUX
    schema.tablespace.temp=TEMP
    standalone.http.port=8888
    standalone.mode=false
    standalone.static.images=/opt/oracle/ords/images/
    standalone.use.https=false
    user.apex.listener.password=BFE2GRPF
    user.apex.restpublic.password=BFE2GRPF
    user.public.password=BFE2GRPF
    user.tablespace.default=SYSAUX
    user.tablespace.temp=TEMP
    sys.user=SYS
    sys.password=br15ban3
  standalone.properties: |
    #Fri May 17 00:45:58 GMT 2019
    jetty.port=8888
    standalone.context.path=/apex
    standalone.doc.root=/opt/oracle/ords/config/ords/standalone/doc_root
    standalone.scheme.do.not.prompt=true
    standalone.static.context.path=/i
    standalone.static.path=/opt/oracle/ords/images/
kind: ConfigMap
metadata:
  name: ordsautoconfig
  labels:
    app: peordshttp
`
Httpconfigmapyml = `
apiVersion: v1
data:
  httpd.conf: |
    ServerRoot "/etc/httpd"
    Include conf.modules.d/*.conf
    User apache
    Group apache
    ServerAdmin root@localhost
    <Directory />
    AllowOverride none
    Require all denied
    </Directory>

    DocumentRoot "/var/www/html"

    <Directory "/var/www">
    AllowOverride None
    Require all granted
    </Directory>

    <Directory "/var/www/html">
    Options Indexes FollowSymLinks

    AllowOverride None

    Require all granted
    </Directory>

    <IfModule dir_module>
    DirectoryIndex index.html
    </IfModule>

    <Files ".ht*">
    Require all denied
    </Files>

    ErrorLog "logs/error_log"
    LogLevel warn
    <IfModule log_config_module>
    LogFormat "%h %l %u %t \"%r\" %>s %b \"%{Referer}i\" \"%{User-Agent}i\"" combined
    LogFormat "%h %l %u %t \"%r\" %>s %b" common

    <IfModule logio_module>
    LogFormat "%h %l %u %t \"%r\" %>s %b \"%{Referer}i\" \"%{User-Agent}i\" %I %O" combinedio
    </IfModule>

    CustomLog "logs/access_log" combined
    </IfModule>

    <IfModule alias_module>
    ScriptAlias /cgi-bin/ "/var/www/cgi-bin/"
    </IfModule>

    <Directory "/var/www/cgi-bin">
    AllowOverride None
    Options None
    Require all granted
    </Directory>

    <IfModule mime_module>
    TypesConfig /etc/mime.types

    AddType application/x-compress .Z
    AddType application/x-gzip .gz .tgz

    AddType text/html .shtml
    AddOutputFilter INCLUDES .shtml
    </IfModule>

    AddDefaultCharset UTF-8

    <IfModule mime_magic_module>
    MIMEMagicFile conf/magic
    </IfModule>

    EnableSendfile on

    IncludeOptional conf.d/*.conf
    Include /etc/httpd/conf/users-define.conf
  users-define.conf: |
    Listen 80
    <VirtualHost *:80>
    Redirect 404 https://livesql-stage.oraclecorp.com/apex/f?p=590:1000
    ErrorDocument 404 https://livesql-stage.oraclecorp.com/apex/f?p=590:1000
    Redirect 403 https://livesql-stage.oraclecorp.com/apex/f?p=590:1000
    ErrorDocument 403 https://livesql-stage.oraclecorp.com/apex/f?p=590:1000


    DocumentRoot "/var/www/html/"
    Alias /i/ "/var/www/html/images/"
    Alias /livesql/ "/var/www/html/images/livesql/"
    Alias /18c "/var/www/html/images/18c/"

    AddType text/xml xbl
    AddType text/x-component htc

    <Directory /var/www/html/>
    AllowOverride none
    Order deny,allow
    Allow from all
    </Directory>

    <Directory /var/www/html/images/>
    Header set X-Frame-Options "deny"
    </Directory>

    <Directory /var/www/html/images/livesql/>
    Header set X-Frame-Options "deny"
    </Directory>

    RewriteEngine On
    RewriteRule ^/$ /apex/f?p=590:1000 [R]
    RewriteRule ^/index.html /apex/f?p=590:1000 [R]
    RewriteRule ^/apex$ /apex/f?p=590:1000 [R]
    RewriteRule ^/apex/$ /apex/f?p=590:1000 [R]

    ProxyPass "/apex" "http://localhost:8888/apex" retry=60
    ProxyPassReverse /apex http://localhost:8888/apex
    ProxyPreserveHost On
    </VirtualHost>
kind: ConfigMap
metadata:
  name: httpautoconfig
  labels:
    app: peordshttp
`
)

