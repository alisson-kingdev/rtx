# RTX - Rotacionador de IP (Cross-Platform)

Projeto modular em Go para rotacao de IP usando proxies publicas e APIs. Suporta **Android (Termux)**, **Windows**, **macOS** e **Linux**.

## Estrutura

```
rtx/
├── cmd/rtx/              # Entry point (Cobra CLI)
│   ├── main.go           # root command + main()
│   └── rotate.go         # rotate subcommand
├── src/
│   ├── controllers/      # Orquestracao
│   │   └── rotation.go
│   ├── services/         # Logica de negocio
│   │   ├── ip.go         # IP publico via APIs
│   │   ├── proxy.go      # Coleta de proxies (3 APIs)
│   │   └── rotation.go   # Logica de rotacao
│   ├── utils/            # Utilitarios
│   │   ├── colors.go     # Cores (fatih/color) + banner ASCII
│   │   └── network.go    # HTTP + extracao de proxies
│   └── models/           # Modelos de dados
│       └── types.go
├── build.sh              # Script de cross-compile
├── Makefile              # Build via Make
├── go.mod
├── go.sum
└── README.md
```

## APIs de Proxy utilizadas

- https://gimmeproxy.com/api/getProxy?curl=true
- https://hproxy.com/api/proxy-list?format=json
- https://api.proxyscraper.com/v1/http.txt

## APIs de IP utilizadas

- https://api.ipify.org
- https://icanhazip.com
- https://ifconfig.me
- https://api.ip.sb/ip
- https://ipinfo.io/ip

## Como usar

### Build para plataforma atual

```bash
go build -o rtx ./cmd/rtx/
./rtx rotate
./rtx rotate --verbose
```

### Cross-compile para todas as plataformas

Usando o script:

```bash
chmod +x build.sh
./build.sh all
```

Usando Make:

```bash
make build-all
```

### Plataformas suportadas

| Plataforma | GOOS | GOARCH | Extensao |
|------------|------|--------|----------|
| Linux x64  | linux | amd64 | (none) |
| Linux ARM64| linux | arm64 | (none) |
| Linux ARM  | linux | arm | (none) |
| macOS Intel| darwin | amd64 | (none) |
| macOS Apple Silicon | darwin | arm64 | (none) |
| Windows x64| windows | amd64 | .exe |
| Android ARM64 (Termux) | android | arm64 | (none) |
| Android ARM (Termux) | android | arm | (none) |

### No Termux (Android)

```bash
pkg install golang
git clone <repo>
cd rtx
./build.sh current
chmod +x rtx
./rtx rotate
```

Ou baixe o binario pre-compilado para `android/arm64` da pasta `build/`.

## Comandos

```bash
rtx rotate          # Executar rotacao de IP
rtx rotate --verbose # Modo verboso

rtx set-proxy <URL>    # Configurar proxy no ambiente do sistema
rtx unset-proxy        # Remover configuracao de proxy do ambiente

rtx --help             # Ajuda
```

### `rtx set-proxy <URL>`

Define variáveis de ambiente de proxy (`http_proxy`, `https_proxy`, `HTTP_PROXY`, `HTTPS_PROXY`, `no_proxy`, `NO_PROXY`) nos arquivos de configuração do shell e, quando aplicável, nas configurações de rede do sistema:

| Plataforma | Arquivos / Comandos utilizados |
|------------|-------------------------------|
| **Linux** | `~/.bashrc`, `~/.bash_profile`, `~/.profile`, `~/.shrc`, `/etc/environment` |
| **macOS** | `networksetup` (Wi-Fi/Ethernet) + `~/.zshrc`, `~/.bash_profile`, `~/.profile` |
| **Windows** | `setx` (variáveis de usuário) + `netsh winhttp` |
| **Android (Termux)** | `~/.bashrc`, `~/.profile` |

**Comportamento inteligente:**
- Se já existir um bloco RTX anterior nos arquivos, ele é **removido** antes de gravar o novo, evitando duplicatas e loops em `.zshrc`, `.bashrc`, `config.fish`, etc.
- O bloco é delimitado por marcadores `# >>> RTX PROXY >>>` e `# <<< RTX PROXY <<<` para detecção segura.

Exemplo:
```bash
rtx set-proxy http://proxy.example.com:8080
rtx set-proxy http://user:pass@proxy:3128
```

### `rtx unset-proxy`

Remove blocos de proxy gerados pelo RTX dos arquivos de configuração do shell e reseta configurações de rede (Windows WinHTTP, macOS networksetup).

Exemplo:
```bash
rtx unset-proxy
```

