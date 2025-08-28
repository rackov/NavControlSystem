├── proto/
│   └── receiver.proto          // <-- прием
│   └── service.proto           // <-- Добавим и общий proto для всех сервисов


protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative service.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative receiver.proto

или
# Находясь в NavControlSystem/
make generate-proto
