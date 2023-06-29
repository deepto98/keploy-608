## Issue 608
### Steps to reproduce the [issue](https://github.com/keploy/keploy/issues/608):
1. Run postgres - `docker-compose up`
2. `export KEPLOY_MODE=record`
3. Run this app :
    ```
    go mod tidy
    go run .
    ```
4. Run Keploy 
5. Send a request to the endpoint:
    ```
    curl --request POST --url http://localhost:8080/test_http_and_sql --header 'content-type: application/json'
    ```