<p align="center">
  <img src="./docs/img/vortex-logo.png" alt="Vortex logo" width="200px" height="200px">
  <h1 align="center">Vortex</h1>
</p>


<p align="center">
  A connector for transferring and transforming data from Kafka to MongoDB.
</p>

<p align="center">
  <img src="https://shields.devops.telekom.de/badge/Made%20with%20%E2%9D%A4%20%20by-%F0%9F%90%BC-blue" alt="Made With Love Badge"/>
  <img src="https://gitlab.devops.telekom.de/dhei/teams/pandora/products/horizon/vortex/badges/develop/coverage.svg" alt="Coverage Badge"/>
  <img src="https://gitlab.devops.telekom.de/dhei/teams/pandora/products/horizon/vortex/-/badges/release.svg" alt="Release Badge"/>
</p>


## Overview
Vortex is a connector for transferring and transforming data from Kafka to MongoDB. It is custom tailored for the data 
produced and processed by Horizon.

## Prerequisites
There's a docker-compose file included in this project which provides a Kafka, Kafka-UI, MongoDB and Zookeeper for local testing.
- A running Kafka broker
- A running MongoDB instance

## Running Vortex
To run vortex simply run `vortex serve` to start processing incoming messages.

## Configuration
Vortex supports configuration via environment variables and/or a configuration file (`config.yml`). The configuration file has to be located in the same directory as the executable and is created by running `vortex init` or `go run . init`.

> **Metrics:** At the moment the metrics are only meant for analyzing the throughput during load tests. 
> They are not meant for monitoring the health of the application. This will probably change in the future.


| Path                       | Variable                          | Type          | Default                   | Description                                                                                                                                                  |
|----------------------------|-----------------------------------|---------------|---------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|
| logLevel                   | VORTEX_LOGLEVEL                   | string        | INFO                      | Sets the overall log-level.                                                                                                                                  |
| metrics.enabled            | VORTEX_METRICS_ENABLED            | bool          | false                     | Enable prometheus metrics on the configured port.                                                                                                            |
| metrics.port               | VORTEX_METRICS_PORT               | int           | 8080                      | The port to use for serving metrics.                                                                                                                         |
| kafka.brokers              | VORTEX_KAFKA_BROKERS              | string (list) | [localhost:9092]          | A list of all brokers.                                                                                                                                       |
| kafka.groupName            | VORTEX_KAFKA_GROUPNAME            | string        | vortex                    | The name of the consumer group used by vortex.                                                                                                               |
| kafka.topics               | VORTEX_KAFKA_TOPICS               | string (list) | [status]                  | A list of all topics to subscribe to.                                                                                                                        |
| kafka.SessionTimeoutSec    | VORTEX_KAFKA_SESSIONTIMEOUTSEC    | int           | 40                        | Max seconds to pass before a forced re-balance.                                                                                                              |
| mongo.url                  | VORTEX_MONGO_URL                  | string        | mongodb://localhost:27017 | The MongoDB url to connect to.                                                                                                                               |
| mongo.database             | VORTEX_MONGO_DATABASE             | string        | horizon                   | The name of the database within MongoDB.                                                                                                                     |
| mongo.collection           | VORTEX_MONGO_COLLECTION           | string        | status                    | The name of the collection within MongoDB.                                                                                                                   |
| mongo.bulkSize             | VORTEX_MONGO_BULKSIZE             | int           | 500                       | The maximal amount per bulk-write (triggers a flush if reached).                                                                                             |
| mongo.flushIntervalSec     | VORTEX_MONGO_FLUSHINTERVALSEC     | int           | 30                        | The amount of seconds between flushes of the bulk buffer.                                                                                                    |
| mongo.writeConcern.writes  | VORTEX_MONGO_WRITECONCERN_WRITES  | int           | 1                         | The amount of writes required for a write to be acknowledged. ([See MongoDB docs](https://www.mongodb.com/docs/manual/reference/write-concern/))             |
| mongo.writeConcern.journal | VORTEX_MONGO_WRITECONCERN_JOURNAL | bool          | false                     | Whether new entries have to be written to disk to be acknowledged or not. ([See MongoDB docs](https://www.mongodb.com/docs/manual/reference/write-concern/)) |

## Code of Conduct
This project has adopted the [Contributor Covenant](https://www.contributor-covenant.org/) in version 2.1 as our code of conduct. Please see the details in our [Code of Conduct](CODE_OF_CONDUCT.md). All contributors must abide by the code of conduct.
By participating in this project, you agree to abide by its [Code of Conduct](./CODE_OF_CONDUCT.md) at all times.

## Licensing
This project follows the [REUSE standard for software licensing](https://reuse.software/).
Each file contains copyright and license information, and license texts can be found in the [LICENSES](./LICENSES) directory. 

For more information visit https://reuse.software/.
For a comprehensive guide on how to use REUSE for licensing in this repository, visit https://telekom.github.io/reuse-template/.   
A brief summary follows below:

The [reuse tool](https://github.com/fsfe/reuse-tool) can be used to verify and establish compliance when new files are added.

For more information on the reuse tool visit https://github.com/fsfe/reuse-tool.

