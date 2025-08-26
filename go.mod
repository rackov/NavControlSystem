module github.com/rackov/NavControlSystem

go 1.23.12

replace github.com/rackov/NavControlSystem/pkg/logger => ./pkg/logger

replace github.com/rackov/NavControlSystem/pkg/tnats => ./pkg/tnats

replace github.com/rackov/NavControlSystem/pkg/models => ./pkg/models

require (
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/nats-io/nats.go v1.45.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/rackov/NavControlSystem/pkg/logger v0.0.0-00010101000000-000000000000 // indirect
	github.com/rackov/NavControlSystem/pkg/tnats v0.0.0-00010101000000-000000000000 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
)
