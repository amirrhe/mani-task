# Many-task

This microservice provides functionality for user authentication, file uploading, and file retrieval.

## Setup

1. **Environment Configuration**
    - To configure the environment variables, copy the `.env.sample` file to `.env` within each microservice directory using the following commands:

    For the **Store Microservice**:
    ```bash
    cd store-microservice
    cp .env.sample .env
    ```

    For the **Retrieval Microservice**:
    ```bash
    cd retreival-microservice
    cp .env.sample .env
    ```

2. **Running the Service**
    
  **Start Docker Services:**

    ```bash
    docker-compose up
    ```

  **Run Store Microservice:**

    After the services are up and running, open a new terminal window or tab, navigate to the `store-microservice` directory, and execute the following command:

    ```bash
    cd store-microservice
    go run main.go
    ```

    Running the above command will start the Store Microservice locally.

  **Note:** Ensure that any necessary dependencies or environment variables are set before running the `main.go` file.
  

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
