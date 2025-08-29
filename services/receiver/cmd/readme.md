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