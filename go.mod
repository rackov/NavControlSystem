module github.com/rackov/NavControlSystem

go 1.24.6

replace github.com/rackov/NavControlSystem/pkg/logger => ./pkg/logger

replace github.com/rackov/NavControlSystem/pkg/tnats => ./pkg/tnats

replace github.com/rackov/NavControlSystem/pkg/models => ./pkg/models

replace github.com/rackov/NavControlSystem/pkg/monitoring => ./pkg/monitoring

require (
	github.com/rackov/NavControlSystem/pkg/logger v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.11.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.23.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.65.0 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/rackov/NavControlSystem/pkg/monitoring v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/sys v0.33.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
