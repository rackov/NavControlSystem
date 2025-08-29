# Сервис receiver
# Для проверки работы сервиса receiver необходимо запустить grpcurl
grpcurl -plaintext -d '{"level": "INFO"}' localhost:50051 proto.LogReader/ReadLogs
# Для установки grpcurl необходимо выполнить команду
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
# Для отклика программы необходимо установить 
go get google.golang.org/grpc/reflection
# сделать изменения  в server.go
reflection.Register(s.grpcServer)

# Установите grpcurl, если еще не установлен: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Чтение всех логов уровня ERROR
grpcurl -plaintext -d '{"level": "ERROR"}' localhost:50051 proto.LogReader/ReadLogs

# Чтение 10 логов за последний час (замените 1698369600 на текущий timestamp минус час)
CURRENT_TS=$(date +%s)
ONE_HOUR_AGO_TS=$((CURRENT_TS - 3600))
grpcurl -plaintext -d "{\"start_date\": $ONE_HOUR_AGO_TS, \"end_date\": $CURRENT_TS, \"limit\": 10}" localhost:50051 proto.LogReader/ReadLogs

# Чтение всех логов с ограничением в 5 строк
grpcurl -plaintext -d '{"limit": 5}' localhost:50051 proto.LogReader/ReadLogs



# 1 .SetLogLevel
grpcurl -plaintext -d '{"level": "DEBUG"}' localhost:50051 proto.ReceiverControl/SetLogLevel

# 2. GetStatus
grpcurl -plaintext -d '{}' localhost:50051 proto.ReceiverControl/GetStatus

# 3. GetActiveConnectionsCount
grpcurl -plaintext -d '{"protocol_name": "ARNAVI"}' localhost:50051 proto.ReceiverControl/GetActiveConnectionsCount

# 4. GetConnectedClients
grpcurl -plaintext -d '{"protocol_name": "ARNAVI"}' localhost:50051 proto.ReceiverControl/GetConnectedClients

# 5. DisconnectClient
grpcurl -plaintext -d '{"protocol_name": "ARNAVI", "client_address": "192.168.1.100:54321"}' localhost:50051 proto.ReceiverControl/DisconnectClient

# 6. OpenPort
grpcurl -plaintext -d '{"id": "a1b2c3d4-e5f6-7890-1234-567890abcdef"}' localhost:50051 proto.ReceiverControl/OpenPort

# 7. ClosePort
grpcurl -plaintext -d '{"id": "a1b2c3d4-e5f6-7890-1234-567890abcdef"}' localhost:50051 proto.ReceiverControl/ClosePort

# 8. AddPort
grpcurl -plaintext -d '{"name": "ARNAVI", "port": 9996}' localhost:50051 proto.ReceiverControl/AddPort

# 9. DeletePort
grpcurl -plaintext -d '{"id": "c3d4e5f6-a7b8-9012-3456-7890abcdef2"}' localhost:50051 proto.ReceiverControl/DeletePort
