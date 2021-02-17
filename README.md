Detecting cloned machine instances
==================================

This is a POC implementation of a service for counting unique machines.

It uses single-use tokens to detect cloned vm instances.

Running
=======

$ docker-compose up

Access http://localhost:8080/status to see the current count of unique machines.

To run multiple client or clone machines, run:

$ docker-compose up --scale client=10 --scale clone=5