# RentalManagement

RentalManagement provides the functionality for the capability [Management of Rentals](https://git.scc.kit.edu/cm-tm/cm-team/projectwork/pse/0-doc-ccs-app-v-2/-/blob/main/pages/capabilities.md) via API endpoints dedicated to individual [use cases](https://git.scc.kit.edu/cm-tm/cm-team/projectwork/pse/0-doc-ccs-app-v-2/-/blob/main/pages/use_case_diagram.md).

For the implementation of the business logic required for the use cases, RentalManagement orchestrates [Car](https://git.scc.kit.edu/cm-tm/cm-team/projectwork/pse/domain/d-carimpl) to access required data.
Therefore, it depends on the private Git repository [CarGoTypes](https://git.scc.kit.edu/cm-tm/cm-team/projectwork/pse/domain/d-cargotypes) to provide mappings for the JSON responses.
Further information on the usage of private Git repositories with go can be found there.

The provided API endpoints of RentalManagement are specified in the [API specification](https://git.scc.kit.edu/cm-tm/cm-team/projectwork/pse/application/rentalmanagementdesign).

## <span style="color: red">[- CORS WARNING -] </span>

The current HELM deployment of this microservice allows requests from all origins. This is a security risk and should be changed in production!
Currently, this is needed for the frontend development to be able to access the API.


## Local Setup
To test RentalManagement locally, you can use the MongoDB Docker Compose setup provided in the `dev` folder.

To do so, execute the following commands:
```bash
cd dev
docker-compose up -d
```

This will start a MongoDB instance on port 27017 with a default user with admin privileges.

After that, start the Go server with the following environment variables:

| Environment Variable                  | Value           | Comment  |
|---------------------------------------|-----------------|----------|
| `MONGODB_DATABASE_HOST`               | localhost       |          |
| `MONGODB_DATABASE_NAME`               | ccsappvp2rental |          |
| `MONGODB_DATABASE_USER`               | root            |          |
| `MONGODB_DATABASE_PASSWORD`           | example         |          |
| `RENTAL_MANAGEMENT_COLLECTION_PREFIX` | someprefix      | optional |
`RENTAL_MANAGEMENT_COLLECTION_PREFIX` provides a prefix to the collection to rule out that different microservice instances
try to use the same collection. This is particularly useful for testing as all tests work on the same database.

## General Setup
You also need to set the environment variable `RM_DOMAIN_SERVER` to the URL of the Car server.
`RM_ALLOW_ORIGINS` may contain a comma-separated list of allowed origins for CORS requests.
Optionally, you can set a timeout for requests to the Car server with `RM_DOMAIN_TIMEOUT`
([number with suffix](https://pkg.go.dev/time#ParseDuration)
"ms" for milliseconds, "s" for seconds, "m" for minutes, "h" for hours)

## Test Setup
The Unit Tests of RentalManagement depend on automatically generated Go mocks.
You need to install [mockgen](https://github.com/golang/mock#installation) to generate them.
After the installation, execute `go generate ./...` in the `src` directory of this project.
The provided API endpoints of RentalManagement are specified in the [API specification](https://git.scc.kit.edu/cm-tm/cm-team/projectwork/pse/application/rentalmanagementdesign).

The test setup loads environment variables from `src/testdata/testdb.env` and ignores the system environment variables.
Also, it overrides `RENTAL_MANAGEMENT_COLLECTION_PREFIX` with a dynamically generated prefix to avoid conflicts with other tests.
