# Docker Compose Notes
## Summary
Our objective was to be able to run the `ATC rebalance` test using Docker instead of BOSH. We discovered that this was very doable with few tweaks to the `suite_test.go` files and a modification of the atc_rebalance_test.go
There will be more work required to be able to pull all the tests into such a format, however, it definitely feels like this is doable.

## Assumptions:
- there should a concourse docker image uploaded to the local machine on which the test runs, the image should have a tag of `topgun`, run `docker build -t topgun -f PATH_TO_DOCKERFILE_IN_TOPGUN_DOCKER_TEST_DIR .` from within the concourse dir referencing the `Dockerfile` provided in the `docker-test` folder

## Problems with the test suite
- We had to add a containerId method to the test suite which may not be compatible with BOSH, however, in Docker land, services can be exposed to the host via LocalHost + Randomly assigned port while in BOSH, there is an virtual network.

## Next Steps
- docker-compose.yml is required for the afterEach & BeforeEach steps of the suite where an explicit deployment is not configured
- docker-compose.yml can become the base yml and have each test specific one override it

# Appendix

## Questions

 1. how to load balance multiple services in the docekr compose file:
    if we use the name of the service directly ... docker-compose automatically load balances services.
 2. giving the service a list of each instance of a service to use.
    - ?????????????????
 3. Having multiple instances and getting their ports:
    - Run docker-compose with `--scale service_name=no_of_instances`
    - in order not to have port issues, make the ports part of the service on exposing container's ports with assigning host ports so docker can assign random ports
    - run `docker-compose port --index=instance_index service_name container_port` in order to get the host assigned port.
 4. Spawning mutiple deployments using the same compose file
    - -p specifies the project to use (by default parent directory name)
    - the command can look like: `docker-compose -f docker-compose.yml -p depl-2 up -d --build --scale web=2`

## Side Notes

- Will need to run docker cleanup commands for every run `docker rm (docker ps -aq)`
- It will be really beneficial to use overrides that kinda resembles the operations files in bosh
- Equivalents:
  - delete-deployment --> docker-compose down
  - deploy --> docker-compose up
  - command will either be
    - docker-compose -f FILE_NAME -p DEPLOYMENT_NAME down
    - docker-compose -f FILE_NAME -p DEPLOYMENT_NAME up --build -d [--extra tags]
