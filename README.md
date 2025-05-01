# Cataloging API

![Go Version](https://img.shields.io/badge/go-1.24-blue)
![MySQL](https://img.shields.io/badge/mysql-4479A1?style=flat&logo=mysql&logoColor=white)

This is a simple Go-based web server application. The application exposes a RESTful API for registering new materials before being created in SAP.

## Table of Contents
1. [Installation](#installation)
2. [Usage](#usage)
3. [Testing](#testing)
4. [Contributing](#contributing)
5. [License](#License)

## Installation

### Prerequsites

Before you begin, ensure you have the following installed:

- Go (version 1.24 or later)
- MySQL (version 8.1.0 or later)

### Steps

1. Clone the repository

```bash
git clone https://github.com/dev-pt-bai/cataloging.git
cd cataloging
```

2. Install dependencies

```bash
go mod tidy
```

3. Set up the environment variables

Copy the `config.json.example` file in `configs` directory to `config.json` and configure the values:

```bash
cp configs/config.json.example configs/config.json
```

Edit `config.json` to include, e.g., your database connection details

```json
{
    "database": {
        "user": "yourdbuser",
        "password": "yourdbpassword",
        "name": "yourdbname"
    }
}
```

This app interacts with Microsoft Graph API. To properly set the related environtmen variables, see the [Microsoft Graph API setup](docs/MSGRAPHAPI.md)

4. Run database migrations

You can run migrations manually or use your favorite tool to apply the migration scripts in the `migrations/` folder. In fact, starting the app will automatically run the migrations too.

To create new migration scripts, see the [migrations guideline](docs/MIGRATIONS.md).

5. Run the server

```bash
go run cmd/app/main.go
```

The application will start at `http://{HOST}:{PORT}`

## Usage

Once the server is running, you can make requests to the API.

### Example API Requests

- **POST** `/login` - Login
- **POST** `/users` - Register a new user
- **GET** `/users/{id}` - Fetch a user by ID
- **PUT** `/users/{id}` - Update a user by ID
- **DELETE** `/users/{id}` - Delete a user by ID

You can use tools like [Postman](https://www.postman.com/) or `curl` to interact with the API.

Example `curl` command:
```bash
curl -X GET http://{HOST}:{PORT}/users
```

### API Endpoints

For more detailed API documentation, see the [API documentation](docs/API.md) or refer to the Swagger documentation.

## Testing

To run tests, simply run:
```bash
go test -cover
```

This will execute all tests in the project.

You can also run tests for specific packages:
```bash
go test internal/handler
```

Tests are located in the respective *_test.go files in each package.

### Test Coverage

To check test coverage:
```bash
go test -cover
```

## Contributing

We welcome contributions! To contribute:

1. Fork the repository
2. Create a new branch: `git checkout -b feature/my-feature`
3. Add your changes: `git add .`
4. Commit your changes: `git commit -m 'add my feature'`
5. Push to the branch: `git push origin feature/my-feature`
6. Create a new Pull Request

For larger changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) for details.