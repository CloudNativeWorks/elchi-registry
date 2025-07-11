# Registry Service

Registry Service, dağıtık controller mimarisi için basit bir service discovery servisidir.

## Özellikler

- **Controller Registry**: Controller'ların kaydedilmesi ve adreslerin paylaşılması
- **Client Location Tracking**: Client'ların hangi controller'a bağlı olduğunun takibi
- **In-Memory Storage**: Hızlı performans için bellek içi veri saklama
- **gRPC API**: Sadece gRPC üzerinden hizmet

## Mimari

```
┌─────────────────┐
│   Load Balancer │
└─────────────────┘
          │
    ┌─────┴─────┐
    │           │
    ▼           ▼
┌──────────┐ ┌──────────┐ ┌──────────┐
│Controller│ │Controller│ │Controller│
│    1     │ │    2     │ │    3     │
└──────────┘ └──────────┘ └──────────┘
    │           │           │
    ▼           ▼           ▼
┌─────────┐ ┌─────────┐ ┌─────────┐
│Client-A │ │Client-B │ │Client-C │
│Client-D │ │Client-E │ │Client-F │
└─────────┘ └─────────┘ └─────────┘
                  │
                  ▼
        ┌─────────────────┐
        │ Registry Service│
        │                 │
        │ - Controller    │
        │   Registry      │
        │ - Client        │
        │   Location      │
        └─────────────────┘
```

## Çalıştırma

```bash
# Varsayılan port ile çalıştır (gRPC: 9090)
go run registry/cmd/main.go

# Özel port ile çalıştır
go run registry/cmd/main.go --grpc-port=9090

# Version bilgisi
go run registry/cmd/main.go --version
```

## gRPC API

Registry service 3 temel operasyon sağlar:

### 1. RegisterController
Controller'ı kaydet:
```protobuf
message ControllerInfo {
    string controller_id = 1;
    string grpc_address = 2;
}

rpc RegisterController(ControllerInfo) returns (ControllerResponse);
```

### 2. GetClientLocation  
Client'ın hangi controller'da olduğunu bul:
```protobuf
message ClientLocationRequest {
    string client_id = 1;
}

message ClientLocationResponse {
    bool found = 1;
    string controller_id = 2;
    string controller_fqdn = 3;
}

rpc GetClientLocation(ClientLocationRequest) returns (ClientLocationResponse);
```

### 3. ForwardCommand
Command forwarding için (gelecekte implement edilecek):
```protobuf
rpc ForwardCommand(ForwardCommandRequest) returns (CommandResponse);
```

## Kullanım Senaryosu

1. **Controller Startup**: Controller ayağa kalktığında kendini registry'e kaydeder
2. **Client Connection**: Client bir controller'a bağlandığında controller bunu registry'e bildirir
3. **Request Routing**: 
   - İstek controller-A'ya gelir
   - Controller-A client_id ile registry'e sorar
   - Registry hangi controller'da olduğunu döner
   - Controller-A o adrese gRPC ile bağlanıp isteği forward eder

## Örnek Kod

```go
// Controller kaydı
client := grpc.Dial("localhost:9090")
registry := NewRegistryClient(client)

// Kendini kaydet
_, err := registry.RegisterController(ctx, &ControllerInfo{
    ControllerID: "ctrl-001",
    GRPCAddress:  "controller1.example.com:50051",
})

// Client location kaydet
registry.SetClientLocation("client-123", "ctrl-001")

// Client location sorgula
resp, err := registry.GetClientLocation(ctx, &ClientLocationRequest{
    ClientID: "client-123",
})
if resp.Found {
    // Forward command to resp.ControllerFQDN
}
``` 