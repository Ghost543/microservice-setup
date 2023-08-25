## Orders Microservice
 - This was developed with an assumption of existence complementary services such as
    - Notification ms
    - Shipping ms
    - Products ms
    - Customers ms
 - Each with its own port but orchestrated to run on the same v-machine, container or deployed under the same vpc
 - Also, each service is expected to be running its own database and assumed to have been built with different programming languages or technologies with an expectation of them communicating to each other.

## Architecture of the MS
Each single ms is a monolithic api only handling one particular task with its own resources.

### Instructions to setup the service
- clone the code from the repo
- run `go mod download`
- Then setup the db from the docker-compose file
- then run `go run main.go`
- And test the endpoints using the `service-test.http` file

### Considerations
- Different teams work on different services
- Each team is free to use what ever technology they deem fit for the service
- Accommodation for the services to communicate to each other if necessary
- Consideration of the CAP rule
- Consideration of scalability on both axes
   
    