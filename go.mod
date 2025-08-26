module github.com/rackov/NavControlSystem

go 1.24.6

replace github.com/rackov/NavControlSystem/pkg/logger => ./pkg/logger

replace github.com/rackov/NavControlSystem/pkg/tnats => ./pkg/tnats

replace github.com/rackov/NavControlSystem/pkg/models => ./pkg/models

require (
	github.com/nats-io/nats.go v1.45.0
	github.com/rackov/NavControlSystem/pkg/logger v0.0.0-00010101000000-000000000000
	github.com/rackov/NavControlSystem/pkg/tnats v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.7.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rackov/NavControlSystem/pkg/models v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)
