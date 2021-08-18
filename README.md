# tiny-echo

tiny-echo is a very small Golang HTTP server which echoes back some information to the client.

CORS has been enabled to avoid any pesky errors when calling the server from web apps.

#### Options

| Name | Description | Default |
|------|-------------|---------|
| name | Name of the server, is included in the response | tiny-echo |
| port | Port number that the server listens on | 80 |
