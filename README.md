# <img alt="Logo" width="80px" src="https://github.com/EnclaveRunner/.github/raw/main/img/enclave-logo.png" style="vertical-align: middle;" /> Enclave Shared Dependencies
> [!WARNING]
> The enclave project is still under heavy development and object to changes. This can include APIs, schemas, interfaces and more. Productive usage is therefore not recommended yet (as long as no stable version is released).


> - Viper for configuration
> - Zap for logging
> - Gin Gonic for HTTP server
> - go-playground/validator for validation


```mermaid
flowchart TB
    %% External Artifacts
    subgraph "External Artifacts"
        Spec["openapi.yml"]:::external
        Proto["health-check.proto"]:::external
    end

    %% Code Generation
    subgraph "Code Generation"
        OAPI["oapi-codegen"]:::process
        Protoc["protoc"]:::process
    end

    %% API Layer
    subgraph "API Layer"
        Gen["api/gen.go\n(Generated, Gin)"]:::generated
        Impl["api/impl.go\n(Handlers)"]:::generated
    end

    %% REST Client
    Client["client/client.gen.go\n(Generated REST Client)"]:::generated

    %% gRPC Health Server
    subgraph "gRPC Health Server"
        G1["proto_gen/health-check.pb.go"]:::generated
        G2["proto_gen/health-check_grpc.pb.go"]:::generated
    end

    %% Middleware Layer
    subgraph "Middleware Layer"
        Authn["middleware/authentication.go\n(Authentication)"]:::custom
        Authz["middleware/authz.go\n(Authorization)"]:::custom
        Log["middleware/zerolog.go\n(Logging)"]:::custom
    end

    %% Business Logic / Domain
    subgraph "Business Logic / Domain"
        Const["auth/const.go"]:::custom
        AuthInit["auth/init.go"]:::custom
        UG["auth/userGroups.go"]:::custom
        RG["auth/resourceGroups.go"]:::custom
        Policies["auth/policies.go"]:::custom
        GM["auth/groupManager.go"]:::custom
        AU["auth/utils.go"]:::custom
        Ptr["utils/ptr.go"]:::custom
        Config["config/config.go"]:::custom
    end

    %% Entry Point
    Main["main.go\n(Bootstrap, Viper)"]:::custom

    %% Infrastructure & Tooling
    subgraph "Infrastructure & Tooling"
        CI1[".github/workflows/ci.yml"]:::infra
        CI2[".github/workflows/release.yml"]:::infra
        CI3[".github/workflows/sync-oapi.yml"]:::infra
        CI4[".github/workflows/testing.yml"]:::infra
        MF["Makefile"]:::infra
        Tools["tools.go"]:::infra
        NixFlake["flake.nix"]:::infra
        NixLock["flake.lock"]:::infra
        Lint[".golangci.yml"]:::infra
    end

    %% Health Probe
    HealthProbe["Health-Check Probe"]:::external

    %% Relationships
    Spec -->|"generate REST code"| OAPI
    Proto -->|"protoc generate gRPC code"| Protoc
    OAPI --> Gen
    OAPI --> Client
    Protoc --> G1
    Protoc --> G2

    Main -->|"loads config"| Config
    Main -->|"start HTTP Server"| Gen
    Main -->|"start HTTP Server"| Impl
    Main -->|"start gRPC Health Server"| G2

    Gen -->|"HTTP Request"| Authn
    Authn -->|"AuthN Check"| Authz
    Authz -->|"AuthZ Check"| Log
    Log -->|"Invoke Handler"| Impl
    Impl -->|"Business Logic Calls"| AuthInit
    Impl --> AuthInit & UG & RG & Policies & GM & AU & Const & Ptr & Config

    HealthProbe -->|"gRPC Probe"| G2

    %% Click Events
    click Spec "https://github.com/enclaverunner/shareddeps/blob/main/openapi.yml"
    click Proto "https://github.com/enclaverunner/shareddeps/blob/main/health-check.proto"
    click Gen "https://github.com/enclaverunner/shareddeps/blob/main/api/gen.go"
    click Impl "https://github.com/enclaverunner/shareddeps/blob/main/api/impl.go"
    click Client "https://github.com/enclaverunner/shareddeps/blob/main/client/client.gen.go"
    click G1 "https://github.com/enclaverunner/shareddeps/blob/main/proto_gen/health-check.pb.go"
    click G2 "https://github.com/enclaverunner/shareddeps/blob/main/proto_gen/health-check_grpc.pb.go"
    click Authn "https://github.com/enclaverunner/shareddeps/blob/main/middleware/authentication.go"
    click Authz "https://github.com/enclaverunner/shareddeps/blob/main/middleware/authz.go"
    click Log "https://github.com/enclaverunner/shareddeps/blob/main/middleware/zerolog.go"
    click AuthInit "https://github.com/enclaverunner/shareddeps/blob/main/auth/init.go"
    click UG "https://github.com/enclaverunner/shareddeps/blob/main/auth/userGroups.go"
    click RG "https://github.com/enclaverunner/shareddeps/blob/main/auth/resourceGroups.go"
    click Policies "https://github.com/enclaverunner/shareddeps/blob/main/auth/policies.go"
    click GM "https://github.com/enclaverunner/shareddeps/blob/main/auth/groupManager.go"
    click AU "https://github.com/enclaverunner/shareddeps/blob/main/auth/utils.go"
    click Const "https://github.com/enclaverunner/shareddeps/blob/main/auth/const.go"
    click Ptr "https://github.com/enclaverunner/shareddeps/blob/main/utils/ptr.go"
    click Config "https://github.com/enclaverunner/shareddeps/blob/main/config/config.go"
    click Main "https://github.com/enclaverunner/shareddeps/blob/main/main.go"
    click CI1 "https://github.com/enclaverunner/shareddeps/blob/main/.github/workflows/ci.yml"
    click CI2 "https://github.com/enclaverunner/shareddeps/blob/main/.github/workflows/release.yml"
    click CI3 "https://github.com/enclaverunner/shareddeps/blob/main/.github/workflows/sync-oapi.yml"
    click CI4 "https://github.com/enclaverunner/shareddeps/blob/main/.github/workflows/testing.yml"
    click MF "https://github.com/enclaverunner/shareddeps/tree/main/Makefile"
    click Tools "https://github.com/enclaverunner/shareddeps/blob/main/tools.go"
    click NixFlake "https://github.com/enclaverunner/shareddeps/blob/main/flake.nix"
    click NixLock "https://github.com/enclaverunner/shareddeps/blob/main/flake.lock"
    click Lint "https://github.com/enclaverunner/shareddeps/blob/main/.golangci.yml"

    %% Styles
    classDef generated fill:#D0E6FF,stroke:#003366,color:#000;
    classDef custom fill:#DFFFD0,stroke:#336600,color:#000;
    classDef external fill:#FFD8A8,stroke:#CC5500,color:#000;
    classDef infra fill:#DDDDDD,stroke:#888888,color:#000;
    classDef process fill:#F0F0F0,stroke:#999999,color:#000;
```
