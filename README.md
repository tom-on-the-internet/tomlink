# Tomlink

tomlink is a nosy redirector that tracks visitors.

It is written in Go, and uses Postgres as a datastore.

This is mostly a learning project.

See it live at [https://tomlink.ca](https://tomlink.ca).

## ENV Variables

"db_url" a postgres connection url (required)
ex: "postgres://postgres:mysecretpassword@localhost:5432/tomlink"

"host" domain this will be hosted on (required)
ex: "https://tomlink.ca" or "http://localhost:3000"

"mock_ip" an ip address to use instead of the client's ip address (optional)
ex: "198.33.22.141"
