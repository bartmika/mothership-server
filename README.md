# mothership-server
The purpose of this application is to provide a gRPC service for storing your internet of things time-series data.

## Motivation

ET (Don't) Phone Home - https://jacquesmattheij.com/et-phone-home/

Google exec says Nest owners should probably warn their guests that their conversations are being recorded - https://www.pulse.ng/bi/tech/google-exec-says-nest-owners-should-probably-warn-their-guests-that-their/z1e1d5n

Google chief: I'd disclose smart speakers before guests enter my home - https://news.ycombinator.com/item?id=21280352

Keep Your Internet Off My Things - https://medium.com/@downey_78309/keep-your-internet-off-my-things-7e35761b66d7

In 2030, You Won't Own Any Gadgets - https://gizmodo.com/in-2030-you-wont-own-any-gadgets-1847176540

## Installation

Get our latest code.

```bash
go install github.com/bartmika/mothership-server@latest
```

## Usage

```text
The purpose of this application is to provide a gRPC service for storing your internet of things time-series data.

Usage:
  mothership-server [flags]
  mothership-server [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  serve       Run the gRPC server
  version     Print the version number

Flags:
  -h, --help   help for mothership-server

Use "mothership-server [command] --help" for more information about a command.
```

### Example
To start the server, run the following command in your **terminal**:

```bash
export MOTHERSHIP_SERVER_DATABASE_URL="postgres://golang:123passwordd@localhost:5432/mothership_db"
export MOTHERSHIP_SERVER_HMAC_SECRET="BLAH_BLAH_PLEASE_CHANGE_THIS_TO_SOMETHING_SUPER_SECRET_BLAH_BLAH"
$GOBIN/mothership-server serve
```

That's it! If everything works, you should see a message saying `Server is running.`.

## Sub-Commands Reference

### ``serve``

**Details:**

```text
Run the gRPC server to allow other services to access this application

Usage:
  mothership-server serve [flags]

Flags:
  -d, --database_url string   The database URL to run this server on (default "postgres://golang:123password@localhost:5432/mothership_db")
  -h, --help                  help for serve
  -s, --hmac_secret string    The secret key to use in this server
  -p, --port int              The port to run this server on (default 50051)
```

**Example:**

```bash
$GOBIN/mothership-server serve -p=50051
```

## Contributing
### Development
If you'd like to setup the project for development. Here are the installation steps:

1. Go to your development folder.

    ```bash
    cd ~/go/src/github.com/bartmika
    ```

2. Clone the repository.

    ```bash
    git clone https://github.com/bartmika/mothership-server.git
    cd mothership-server
    ```

3. Install the package dependencies

    ```bash
    go mod tidy
    ```

4. In your **terminal**, make sure we export our path (if you haven’t done this before) by writing the following:

    ```bash
    export PATH="$PATH:$(go env GOPATH)/bin"
    ```

5. Run the following to generate our new gRPC interface. Please note in your development, if you make any changes to the gRPC service definition then you'll need to rerun the following:

    ```bash
    protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/mothership.proto
    ```

6. You are now ready to start the server and begin contributing! (Don't forget to apply the environment variables as well)

    ```bash
    go run main.go serve
    ```

### Quality Assurance

Found a bug? Need Help? Please create an [issue](https://github.com/bartmika/mothership-server/issues).


## License

[**ISC License**](LICENSE) © Bartlomiej Mika
