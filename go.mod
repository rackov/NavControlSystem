module github.com/rackov/NavControlSystem

go 1.24.6

replace github.com/rackov/NavControlSystem/pkg/logger => ./pkg/logger

replace github.com/rackov/NavControlSystem/pkg/tnats => ./pkg/tnats

require (
	github.com/nats-io/nats.go v1.45.0
	github.com/rackov/NavControlSystem/pkg/logger v0.0.0-00010101000000-000000000000
	github.com/rackov/NavControlSystem/pkg/tnats v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
)
