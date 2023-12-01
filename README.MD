# Many-task

This microservice provides functionality for user authentication, file uploading, and file retrieval.

## Setup

1. **Environment Configuration**
    - To configure the environment variables, copy the `.env.sample` file to `.env` within each microservice directory using the following commands:

    For the **Store Microservice**:
    ```bash
    cd store_microservice
    cp .env.sample .env
    ```

    For the **Retrieval Microservice**:
    ```bash
    cd retrieval_microservice
    cp .env.sample .env
    ```

2. **Running the Service**
    - After configuring the environment, start the service by running the following command in the root directory where `docker-compose.yml` is located:
    ```bash
    docker-compose up
    ```
    
    This command will initiate the service using Docker Compose, ensuring all necessary components are up and running.

## Endpoints

### User Authentication

- **Register User**
  - Method: `POST`
  - Endpoint: `/api/v1/user/register`
  - Request Body:
    ```json
    {
        "username": "example",
        "email": "example@example.com",
        "password": "yourpassword",
        "first_name": "John",
        "last_name": "Doe"
    }
    ```

- **Login User**
  - Method: `POST`
  - Endpoint: `/api/v1/user/login`
  - Request Body:
    ```json
    {
        "identifier": "example@example.com",
        "password": "yourpassword"
    }
    ```
    *Note: `identifier` can be an email or username.*

### File Handling

- **Upload File**
  - Method: `POST`
  - Endpoint: `/api/v1/file`
  - Authentication: JWT Token required.
  - Request: Form data with a field `file`, `tag` and `type` .

- **Get File**
  - Method: `GET`
  - Endpoint: `/api/v1/file`
  - Authentication: JWT Token required.
  - Query Params: `tags` or `name` for filtering files.

- **Retrieve File Names**
  - Method: `GET`
  - Endpoint: `/api/v1/file/names`
  - Authentication: JWT Token required.
  - Retrieves file names based on provided tags or name.

## Documentation

- Postman Collection:
  - Access the Postman collection in the `docs` folder for detailed API usage.

## Running Tests

- To execute Golang tests for the Retrieval Microservice:
  ```bash
  cd retrieval_microservice_directory
  go test ./...
  ```