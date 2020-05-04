# EasyCab Backend System

_An implementation of a functional Taxi/Cab ordering Backend System in Go_

**General Usage Instructions:**

- It's advised to use reverse proxy to manage the traffic between servers.
- It's recommended to use different containers/hosts for each server within same subnet.
- It's also recommended to have Mongodb running
- Redis can be used to send invoices in emails or different channels.

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

2.  Host each server on same subnet for better performance.
3.  Use a reverse proxy to manage the traffics,for example NGINX.
4.  Start a new Mongodb Server/Deamon:

    <code>
    ~/mongodb/bin/mongod --dbpath /data/db
    </code>

5.  Insert Demo JSON file to Mongo DB Collection:

    <code>
    ~/mongodb/bin/mongoimport --db DB_NAME --collection COLLECTION_NAME --drop --file= /LOCATION/OF/YOUR/JSON/DB/FILE.json
    </code>

6.  Create 2D Index such as:

    <code>
    db.COLLECTION_NAME.ensureIndex({NAME_OF_LOCATION_FIELD:"2dsphere"})
    </code>

7.  Make sure that if you create HTTP request to include in header the below:

    <code>
    Content-Type application/vnd.api+json
    Accept application/vnd.api+json
    Authorization Bearer xxxxxxx
    </code>

> Feel free to PR and Review
