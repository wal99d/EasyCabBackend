# EasyCab Backend System

_An implementation of a functional Taxi/Cab ordering Backend System in Go_

**General Usage Instructions:**

- It's advised to use reverse proxy to manage the traffic between servers.
- It's recommended to use different containers/host for each server within same subnet.
- It's also recommended to have Mongodb running
- Redis can be used to send invoices in emails.

**Technologies Used:**

- Rest JSON API
- GeoJSON
- JWT
- Mongodb

**How-To:**

1.  Compile the src from servers directory:

    > go build servers/biller/main.go biller
    > go build servers/dispatcher/main.go disptacher
    > go build servers/push/main.go push
    > go build servers/profile/main.go profile

2.  Host each server on same subnet for better performance
3.  Use a reverse proxy to manage the traffics, below example for using NGINX:
    <code>
    user nobody;

    worker_processes 4;

    events {
    worker_connections 1024;
    }

    http {
    include mime.types;
    default_type application/octet-stream;

         sendfile        on;
         keepalive_timeout  65;

         server {
             listen       30001;
             server_name  localhost;

             location / {
                 root   /Library/WebServer/Documents;
                 autoindex on;
             }

         location ~* \.(eot|ttf|woff)$ {
                 add_header Access-Control-Allow-Origin *;
         }

    }

     </code>

4.  Start new Mongodb Server/Deamon
    ~/mongodb/bin/mongod --dbpath /data/db
5.  Insert Demo JSON file to Mongo DB Collection
    ~/mongodb/bin/mongoimport --db DB_NAME --collection COLLECTION_NAME --drop --file= /LOCATION/OF/YOUR/JSON/DB/FILE.json
6.  Create 2D Index
    db.COLLECTION_NAME.ensureIndex({NAME_OF_LOCATION_FIELD:"2dsphere"})
7.  Make sure that if you create HTTP request to include in header the below:
    Content-Type application/vnd.api+json
    Accept application/vnd.api+json
    Authorization Bearer xxxxxxx

> Feel free to PR and Review
